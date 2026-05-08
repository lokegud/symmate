package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/loke/symmate/internal/llm"
)

// Turn is one exchange in the sim transcript.
type Turn struct {
	TurnNum  int    `json:"turn"`
	Role     string `json:"role"` // "persona" | "symmate"
	Persona  string `json:"persona,omitempty"`
	Content  string `json:"content"`
	Duration string `json:"duration_ms,omitempty"`
}

// Runner orchestrates a sim conversation.
type Runner struct {
	TargetURL  string // e.g. http://localhost:7842
	MaxTurns   int
	Model      string // model name for persona agent (passed to stak)
	hermes     *llm.HermesProvider
	client     *http.Client
}

func NewRunner(targetURL string, maxTurns int, model string, hermes *llm.HermesProvider) *Runner {
	return &Runner{
		TargetURL: targetURL,
		MaxTurns:  maxTurns,
		Model:     model,
		hermes:    hermes,
		client:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// Run executes the full scenario conversation and returns the transcript.
// For single-persona scenarios the conversation is persona ↔ symmate.
// For two-persona scenarios (h2h type) both personas talk to the same symmate target.
func (r *Runner) Run(ctx context.Context, sc *Scenario, personas map[string]*Persona) ([]Turn, error) {
	switch sc.Type {
	case "h2h":
		return r.runH2H(ctx, sc, personas)
	default:
		return r.runSingle(ctx, sc, personas)
	}
}

// runSingle: one persona vs one symmate instance.
func (r *Runner) runSingle(ctx context.Context, sc *Scenario, personas map[string]*Persona) ([]Turn, error) {
	if len(sc.Personas) == 0 {
		return nil, fmt.Errorf("scenario has no personas")
	}
	pr := sc.Personas[0]
	persona, ok := personas[pr.Role]
	if !ok {
		return nil, fmt.Errorf("persona %q not loaded", pr.Role)
	}

	sysPrompt := persona.SystemPrompt(pr.Role, sc.Goal)
	// persona agent history (system + turns)
	agentMsgs := []llm.Message{
		{Role: "system", Content: sysPrompt},
	}

	var transcript []Turn
	// opening message from scenario body (first blockquote or first turn)
	openingMsg := extractOpeningMessage(sc.Body)

	var lastSymmateReply string
	for turn := 0; turn < r.MaxTurns; turn++ {
		// --- Persona turn ---
		var personaMsg string
		if turn == 0 && openingMsg != "" {
			personaMsg = openingMsg
		} else {
			// Ask persona agent to generate next message
			if lastSymmateReply != "" {
				agentMsgs = append(agentMsgs, llm.Message{
					Role:    "user",
					Content: fmt.Sprintf("The other party said:\n%s\n\nYour reply:", lastSymmateReply),
				})
			} else {
				agentMsgs = append(agentMsgs, llm.Message{
					Role:    "user",
					Content: "Start the conversation with your opening message.",
				})
			}

				t0 := time.Now()
			reply, err := r.hermes.Complete(ctx, r.Model, agentMsgs)
			if err != nil {
				return transcript, fmt.Errorf("persona agent error turn %d: %w", turn, err)
			}
			elapsed := time.Since(t0).Milliseconds()
			reply = strings.TrimSpace(reply)
			fmt.Printf("[sim] turn %d persona  %5dms\n", turn+1, elapsed)

			// Append assistant turn to persona agent history
			agentMsgs = append(agentMsgs, llm.Message{Role: "assistant", Content: reply})

			if reply == "[END]" || strings.HasPrefix(reply, "[END]") {
				break
			}
			personaMsg = reply
		}

		transcript = append(transcript, Turn{
			TurnNum: turn + 1,
			Role:    "persona",
			Persona: pr.Persona,
			Content: personaMsg,
		})

		// --- Symmate turn ---
		t0 := time.Now()
		symmateReply, err := r.chatWithSymmate(ctx, personaMsg)
		if err != nil {
			return transcript, fmt.Errorf("symmate error turn %d: %w", turn, err)
		}
		elapsed := time.Since(t0).Milliseconds()
		fmt.Printf("[sim] turn %d symmate  %5dms\n", turn+1, elapsed)

		transcript = append(transcript, Turn{
			TurnNum:  turn + 1,
			Role:     "symmate",
			Content:  symmateReply,
			Duration: fmt.Sprintf("%d", elapsed),
		})
		lastSymmateReply = symmateReply
	}

	return transcript, nil
}

// runH2H: two personas mediated by the same symmate target.
// Both personas send to the same conversation thread (same visitor handle).
// The mediator (symmate) sees interleaved messages from both.
func (r *Runner) runH2H(ctx context.Context, sc *Scenario, personas map[string]*Persona) ([]Turn, error) {
	if len(sc.Personas) < 2 {
		return nil, fmt.Errorf("h2h scenario needs at least 2 personas")
	}

	type party struct {
		role    string
		persona *Persona
		msgs    []llm.Message // persona agent history
	}

	var parties [2]party
	for i, pr := range sc.Personas[:2] {
		p, ok := personas[pr.Role]
		if !ok {
			return nil, fmt.Errorf("persona %q not loaded", pr.Role)
		}
		parties[i] = party{
			role:    pr.Role,
			persona: p,
			msgs: []llm.Message{
				{Role: "system", Content: p.SystemPrompt(pr.Role, sc.Goal)},
			},
		}
	}

	var transcript []Turn
	lastSymmateToParty := [2]string{}

	// Seeded context from scenario body
	seed := extractH2HSeed(sc.Body)
	if len(seed) >= 2 {
		for i, line := range seed[:2] {
			parties[i].msgs = append(parties[i].msgs, llm.Message{
				Role:    "assistant",
				Content: line,
			})
			transcript = append(transcript, Turn{
				TurnNum: 1,
				Role:    "persona",
				Persona: parties[i].role,
				Content: line,
			})
		}
	}

	for turn := 1; turn < r.MaxTurns; turn++ {
		for i := range parties {
			p := &parties[i]

			prompt := "Continue the conversation."
			if lastSymmateToParty[i] != "" {
				prompt = fmt.Sprintf("The mediator said to you privately:\n%s\n\nThe other party's last message is what you saw. Your reply:", lastSymmateToParty[i])
			}
			p.msgs = append(p.msgs, llm.Message{Role: "user", Content: prompt})

			reply, err := r.hermes.Complete(ctx, r.Model,p.msgs)
			if err != nil {
				return transcript, fmt.Errorf("persona agent %s error: %w", p.role, err)
			}
			reply = strings.TrimSpace(reply)
			p.msgs = append(p.msgs, llm.Message{Role: "assistant", Content: reply})

			if reply == "[END]" || strings.HasPrefix(reply, "[END]") {
				goto done
			}

			transcript = append(transcript, Turn{
				TurnNum: turn + 1,
				Role:    "persona",
				Persona: p.role,
				Content: reply,
			})

			// Send to mediator (symmate) with visitor handle to distinguish party
			symmateReply, err := r.chatWithSymmateHandle(ctx, reply, p.role)
			if err != nil {
				return transcript, fmt.Errorf("symmate error for %s: %w", p.role, err)
			}

			transcript = append(transcript, Turn{
				TurnNum: turn + 1,
				Role:    "symmate",
				Persona: p.role + "_response",
				Content: symmateReply,
			})
			lastSymmateToParty[i] = symmateReply
		}
	}

done:
	return transcript, nil
}

// chatWithSymmate sends one message to the Symmate visitor API and collects the full SSE reply.
func (r *Runner) chatWithSymmate(ctx context.Context, message string) (string, error) {
	return r.chatWithSymmateHandle(ctx, message, "")
}

func (r *Runner) chatWithSymmateHandle(ctx context.Context, message, handle string) (string, error) {
	reqBody := map[string]string{"message": message}
	if handle != "" {
		reqBody["visitor_handle"] = "sim-" + handle
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.TargetURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("symmate HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("symmate returned %d", resp.StatusCode)
	}

	return readSSE(resp)
}

// readSSE reads a Symmate SSE response and returns the concatenated content.
// Format: lines starting with "data: <text>", ending with "data: [DONE]".
func readSSE(resp *http.Response) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := line[6:]
		if payload == "[DONE]" {
			break
		}
		sb.WriteString(payload)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(sb.String()), nil
}

// extractOpeningMessage pulls the first customer/attacker turn from the scenario body.
// Accepts blockquote format ("> text") and bracket format ("[CUSTOMER]: text", "[ATTACKER]: text").
func extractOpeningMessage(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "> ") {
			return strings.TrimPrefix(line, "> ")
		}
		if strings.HasPrefix(line, ">") && len(line) > 1 {
			return strings.TrimPrefix(line, ">")
		}
		// "[CUSTOMER]: ..." or "[ATTACKER]: ..."
		if strings.HasPrefix(line, "[") {
			closeBracket := strings.Index(line, "]:")
			if closeBracket > 0 {
				return strings.TrimSpace(line[closeBracket+2:])
			}
		}
	}
	return ""
}

// extractH2HSeed returns the first 2+ quoted lines from a conversation history block.
func extractH2HSeed(body string) []string {
	var lines []string
	inHistory := false
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, "# Conversation history") || strings.Contains(line, "# Conversation transcript") {
			inHistory = true
			continue
		}
		if inHistory && strings.HasPrefix(strings.TrimSpace(line), "#") {
			break
		}
		if inHistory {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "**Party") || strings.HasPrefix(line, "**party") {
				// extract quoted content after the label
				idx := strings.Index(line, "** ")
				if idx >= 0 {
					msg := strings.TrimSpace(line[idx+3:])
					msg = strings.Trim(msg, `"`)
					if msg != "" {
						lines = append(lines, msg)
					}
				}
			}
		}
	}
	return lines
}
