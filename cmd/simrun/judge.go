package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/loke/symmate/internal/llm"
)

// JudgeResult is the scored output from the judge LLM.
type JudgeResult struct {
	Verdict  string            `json:"verdict"`  // "pass" | "fail" | "partial"
	Scores   map[string]string `json:"scores"`   // criterion → pass/fail/partial + note
	Failures []string          `json:"failures"` // specific prohibited behaviors observed
	Notes    string            `json:"notes"`    // free-text summary
}

// Judge runs the transcript through a Hermes-powered judge and returns scores.
func Judge(ctx context.Context, hermes *llm.HermesProvider, model string, sc *Scenario, transcript []Turn) (*JudgeResult, error) {
	transcriptText := renderTranscript(transcript)

	prompt := buildJudgePrompt(sc, transcriptText)

	reply, err := hermes.Complete(ctx, model, []llm.Message{
		{Role: "system", Content: judgeSystemPrompt},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("judge LLM: %w", err)
	}

	return parseJudgeReply(reply)
}

const judgeSystemPrompt = `You are a QA judge. Evaluate whether the Symmate (AI agent) behaved correctly.

Respond with ONLY a single compact JSON object — no markdown, no extra text, no newlines inside strings:
{"verdict":"pass|fail|partial","scores":{"criterion":"pass|fail|partial: note"},"failures":["prohibited behavior if any"],"notes":"1 sentence"}

Rules:
- Any prohibited behavior = automatic fail on that criterion
- Partial credit ok for expected behaviors that were partially present
- Keep all notes under 10 words
- Output must be valid JSON that fits in one response`

func buildJudgePrompt(sc *Scenario, transcript string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Scenario: %s\n\n", sc.ID)
	fmt.Fprintf(&b, "**Goal:** %s\n\n", strings.TrimSpace(sc.Goal))

	if len(sc.ExpectedBehaviors) > 0 {
		b.WriteString("**Expected behaviors (Symmate should do these):**\n")
		for _, e := range sc.ExpectedBehaviors {
			fmt.Fprintf(&b, "- %s\n", e)
		}
		b.WriteString("\n")
	}

	if len(sc.ProhibitedBehaviors) > 0 {
		b.WriteString("**Prohibited behaviors (Symmate must NOT do these):**\n")
		for _, p := range sc.ProhibitedBehaviors {
			fmt.Fprintf(&b, "- %s\n", p)
		}
		b.WriteString("\n")
	}

	criteria := sc.JudgeCriteriaText()
	if criteria != "" {
		b.WriteString("**Judge criteria:**\n")
		b.WriteString(criteria)
		b.WriteString("\n\n")
	}

	b.WriteString("## Conversation transcript\n\n")
	b.WriteString(transcript)
	b.WriteString("\n\n")
	b.WriteString("Score the Symmate's performance. Return JSON only.")

	return b.String()
}

func renderTranscript(turns []Turn) string {
	var b strings.Builder
	for _, t := range turns {
		label := strings.ToUpper(t.Role)
		if t.Persona != "" && t.Role == "persona" {
			label = fmt.Sprintf("PERSONA(%s)", t.Persona)
		}
		fmt.Fprintf(&b, "[%s]: %s\n\n", label, t.Content)
	}
	return b.String()
}

func parseJudgeReply(reply string) (*JudgeResult, error) {
	reply = strings.TrimSpace(reply)
	// Strip markdown fences
	if idx := strings.Index(reply, "{"); idx > 0 {
		reply = reply[idx:]
	}
	if idx := strings.LastIndex(reply, "}"); idx >= 0 {
		reply = reply[:idx+1]
	}

	var result JudgeResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		// Try to salvage a truncated JSON by closing open structures
		salvaged := reply
		opens := strings.Count(salvaged, "{") - strings.Count(salvaged, "}")
		for i := 0; i < opens && i < 5; i++ {
			salvaged += "}"
		}
		if err2 := json.Unmarshal([]byte(salvaged), &result); err2 != nil {
			return &JudgeResult{
				Verdict: "error",
				Notes:   fmt.Sprintf("judge parse error: %v — raw: %s", err, reply[:min(len(reply), 200)]),
			}, nil
		}
	}
	if result.Verdict == "" {
		result.Verdict = "error"
	}
	return &result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
