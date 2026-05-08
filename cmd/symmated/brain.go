package main

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/loke/symmate/internal/aeo"
	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/personality"
	"github.com/loke/symmate/internal/tool"
	"github.com/loke/symmate/internal/visitor"
	"github.com/rs/zerolog/log"
)

type Brain struct {
	loader   *personality.Loader
	router   *llm.Router
	firewall *firewall.Firewall
	grants   []firewall.Grant
	hostDir  string
	tools    *tool.Registry
	toolTags []string // allowed tag prefixes; nil = all tools
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

	if ownerMode {
		sysPrompt += ownerSystemPromptSuffix
	}

	var toolDefs []llm.ToolDef
	if b.tools != nil {
		toolDefs = b.tools.Defs(b.toolTags...)
	}
	if b.tools != nil && len(toolDefs) > 0 {
		sysPrompt += b.tools.Skills(b.toolTags...)
	}

	llmMsgs := make([]llm.Message, 0, len(msgs)+1)
	if sysPrompt != "" {
		llmMsgs = append(llmMsgs, llm.Message{Role: "system", Content: sysPrompt})
	}
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, llm.Message{Role: m.Role, Content: m.Content})
	}

	if ownerMode {
		// Non-streaming: lets us strip directives before the owner sees them.
		ch := make(chan string, 1)
		go func() {
			defer close(ch)
			llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()
			var reply string
			var err error
			if len(toolDefs) > 0 {
				reply, err = b.router.RouteWithTools(llmCtx, tag, llmMsgs, toolDefs, b.tools.Executor(b.toolTags...))
			} else {
				reply, err = b.router.Route(llmCtx, tag, llmMsgs)
			}
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
		if len(toolDefs) > 0 {
			reply, err := b.router.RouteWithTools(llmCtx, tag, llmMsgs, toolDefs, b.tools.Executor(b.toolTags...))
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					ch <- "(error: " + err.Error() + ")"
				}
				return
			}
			select {
			case <-ctx.Done():
			case ch <- reply:
			}
			return
		}
		if err := b.router.RouteStream(llmCtx, tag, llmMsgs, ch); err != nil {
			select {
			case <-ctx.Done():
			default:
				ch <- "(error: " + err.Error() + ")"
			}
		}
	}()
	return ch, nil
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
