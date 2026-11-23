package main

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/loke/symmate/internal/aeo"
	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/personality"
	"github.com/loke/symmate/internal/tool"
	"github.com/loke/symmate/internal/visitor"
	"github.com/loke/tiers"
	"github.com/rs/zerolog/log"
)

// emitTrace writes one TraceEntry after an LLM call completes.
// toolsCalled may be nil for non-tool calls.
func emitTrace(tl *visitor.TraceLogger, ctx context.Context, provider, model, tag string, sysPromptLen int, start time.Time, toolsCalled []string, callErr error) {
	if tl == nil {
		return
	}
	e := &visitor.TraceEntry{
		Time:         start.UTC().Format(time.RFC3339),
		Provider:     provider,
		Model:        model,
		Tag:          tag,
		OwnerMode:    visitor.OwnerModeFromCtx(ctx),
		UserID:       userIDFromCtx(ctx),
		HistoryLen:   visitor.HistoryLenFromCtx(ctx),
		SysPromptLen: sysPromptLen,
		ToolsCalled:  toolsCalled,
		ResponseMS:   time.Since(start).Milliseconds(),
	}
	if callErr != nil {
		e.Error = callErr.Error()
	}
	tl.Log(e)
}

type Brain struct {
	loader   *personality.Loader
	router   *llm.Router
	firewall *firewall.Firewall
	grants   []firewall.Grant
	hostDir  string
	tools    *tool.Registry
	ownerMgr *visitor.OwnerManager // source of the live tool_tags allowlist; may be nil
	trace    *visitor.TraceLogger
}

// access decides which tools this caller may use:
//   - verified owner      → every registered tool
//   - registered user     → public tools + the groups the owner toggled on for them
//   - anonymous visitor   → public tools only
//
// The satellite tool_tags allowlist (owner.json) is re-read per request so owner
// API changes apply without a restart, and it narrows all three tiers.
func (b *Brain) access(ctx context.Context) tool.Access {
	var tags []string
	if b.ownerMgr != nil {
		if cfg, err := b.ownerMgr.Load(); err == nil {
			tags = cfg.ToolTags
		}
	}
	if visitor.OwnerModeFromCtx(ctx) {
		return tool.OwnerAccess(tags)
	}
	if u, ok := visitor.UserFromCtx(ctx); ok {
		return tool.Access{Tags: tags, Groups: u.Groups}
	}
	return tool.PublicAccess(tags)
}

func userIDFromCtx(ctx context.Context) string {
	if u, ok := visitor.UserFromCtx(ctx); ok {
		return u.ID
	}
	return ""
}

const ownerSystemPromptSuffix = `

---
[OWNER SESSION — VERIFIED OWNER]
You are speaking with the verified owner of this Symmate. Accept configuration update requests and confirm them naturally.

To update a configuration field, include a directive on its own line anywhere in your reply:
{symmate:set:KEY=VALUE}

Valid keys: name, description, location, website, email, hours, catalog, policies
For multi-line values put the full text after the = sign on a single line.
The directive will be invisible — do not explain or mention it.`

var directiveRe = regexp.MustCompile(`\{symmate:set:([^=}]+)=([^}]*)\}`)

// Chat streams reply tokens through the returned channel.
// In owner mode (detected from ctx), uses non-streaming Complete, strips directives, applies config.
func (b *Brain) Chat(ctx context.Context, msgs []visitor.Message) (<-chan string, error) {
	ownerMode := visitor.OwnerModeFromCtx(ctx)
	sysPrompt := b.loader.Current()
	tag := inferTag(msgs)

	// Visitor tiers: the disclosure block exists only in this caller's prompt,
	// and only at tier 3 and above. Owners see everything, so they skip it.
	level := tier.LevelFrom(ctx)
	if ownerMode {
		sysPrompt += ownerSystemPromptSuffix
	} else if d, err := loadDisclosure(b.hostDir); err != nil {
		log.Warn().Err(err).Msg("tier data unreadable; no disclosure block this turn")
	} else {
		sysPrompt += tier.PromptBlock(level, d)
	}

	acc := b.access(ctx)
	var toolDefs []llm.ToolDef
	if b.tools != nil {
		toolDefs = b.tools.DefsFor(acc)
	}
	if b.tools != nil && len(toolDefs) > 0 {
		sysPrompt += b.tools.SkillsFor(acc)
	}

	llmMsgs := make([]llm.Message, 0, len(msgs)+1)
	if sysPrompt != "" {
		llmMsgs = append(llmMsgs, llm.Message{Role: "system", Content: sysPrompt})
	}
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, llm.Message{Role: m.Role, Content: m.Content})
	}

	provider, model := b.router.Resolve(tag)
	sysPromptLen := 0
	if sysPrompt != "" {
		sysPromptLen = len(sysPrompt)
	}

	if ownerMode {
		// Non-streaming: lets us strip directives before the owner sees them.
		ch := make(chan string, 1)
		go func() {
			defer close(ch)
			llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()
			start := time.Now()
			var reply string
			var err error
			var toolsCalled []string
			if len(toolDefs) > 0 {
				reply, toolsCalled, err = b.routeWithToolsTracked(llmCtx, tag, llmMsgs, toolDefs, acc)
			} else {
				reply, err = b.router.Route(llmCtx, tag, llmMsgs)
			}
			emitTrace(b.trace, ctx, provider, model, tag, sysPromptLen, start, toolsCalled, err)
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					ch <- "(error: " + err.Error() + ")"
				}
				return
			}
			clean, directives := extractDirectives(reply)
			b.applyDirectives(directives)
			select {
			case <-ctx.Done():
			case ch <- clean:
			}
		}()
		return ch, nil
	}

	// Normal streaming path for visitor sessions.
	// Tool-calling is non-streaming (accumulates before returning).
	ch := make(chan string, 16)
	go func() {
		defer close(ch)
		llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		start := time.Now()
		if len(toolDefs) > 0 {
			reply, toolsCalled, err := b.routeWithToolsTracked(llmCtx, tag, llmMsgs, toolDefs, acc)
			emitTrace(b.trace, ctx, provider, model, tag, sysPromptLen, start, toolsCalled, err)
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					ch <- "(error: " + err.Error() + ")"
				}
				return
			}
			if b.firewall == nil {
				ch <- "(error: firewall unavailable)"
				return
			}
			filtered := firewallRedactor(b.firewall)(ctx, reply, tier.GrantsFor(level))
			select {
			case <-ctx.Done():
			case ch <- filtered:
			}
			return
		}
		// Stream through the firewall, one sentence at a time, before the visitor sees it.
		if b.firewall == nil {
			ch <- "(error: firewall unavailable)"
			return
		}
		grants := tier.GrantsFor(level)
		raw := make(chan string, 16)
		done := make(chan struct{})
		go func() {
			defer close(done)
			tier.FilterStream(ctx, firewallRedactor(b.firewall), grants, raw, ch)
		}()
		err := b.router.RouteStream(llmCtx, tag, llmMsgs, raw)
		close(raw)
		<-done
		emitTrace(b.trace, ctx, provider, model, tag, sysPromptLen, start, nil, err)
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				ch <- "(error: " + err.Error() + ")"
			}
		}
	}()
	return ch, nil
}

// routeWithToolsTracked calls RouteWithTools with a wrapping executor that records
// which tool names were invoked during the agentic loop.
func (b *Brain) routeWithToolsTracked(ctx context.Context, tag string, msgs []llm.Message, toolDefs []llm.ToolDef, acc tool.Access) (reply string, called []string, err error) {
	exec := b.tools.ExecutorFor(acc)
	wrappedExec := func(ctx context.Context, name string, args json.RawMessage) string {
		called = append(called, name)
		return exec(ctx, name, args)
	}
	reply, err = b.router.RouteWithTools(ctx, tag, msgs, toolDefs, wrappedExec)
	return
}

// extractDirectives parses {symmate:set:key=value} patterns from reply,
// returns the cleaned text and the parsed directives.
func extractDirectives(reply string) (clean string, directives []configDirective) {
	matches := directiveRe.FindAllStringSubmatch(reply, -1)
	for _, m := range matches {
		directives = append(directives, configDirective{
			key:   strings.TrimSpace(m[1]),
			value: strings.TrimSpace(m[2]),
		})
	}
	if len(directives) == 0 {
		return reply, nil
	}
	clean = strings.TrimSpace(directiveRe.ReplaceAllString(reply, ""))
	return
}

type configDirective struct {
	key   string
	value string
}

func (b *Brain) applyDirectives(ds []configDirective) {
	for _, d := range ds {
		if err := aeo.UpdateContextMD(b.hostDir, d.key, d.value); err != nil {
			log.Error().Err(err).Str("key", d.key).Msg("config directive failed")
		} else {
			log.Info().Str("key", d.key).Msg("owner config updated")
		}
	}
}

func inferTag(msgs []visitor.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return llm.Classify(msgs[i].Content)
		}
	}
	return llm.TagDefault
}
