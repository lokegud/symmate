package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/loke/symmate/internal/llm"
)

func main() {
	var (
		scenarioPath string
		personaDir   string
		targetURL    string
		outputDir    string
		maxTurns     int
		model        string
		noJudge      bool
	)

	flag.StringVar(&scenarioPath, "scenario", "", "path to scenario markdown file (required)")
	flag.StringVar(&personaDir, "persona-dir", "/mnt/KB/sim/personas", "directory containing persona YAML files")
	flag.StringVar(&targetURL, "target", "http://localhost:7842", "Symmate visitor API base URL")
	flag.StringVar(&outputDir, "output-dir", "/mnt/KB/sim/runs", "directory for run reports")
	flag.IntVar(&maxTurns, "max-turns", 10, "maximum conversation turns")
	flag.StringVar(&model, "model", "anthropic-claude-haiku-4-5", "LLM model for persona agent and judge (via stak)")
	flag.BoolVar(&noJudge, "no-judge", false, "skip judge scoring")
	flag.Parse()

	if scenarioPath == "" {
		fmt.Fprintln(os.Stderr, "error: --scenario is required")
		flag.Usage()
		os.Exit(1)
	}

	// Allow bare scenario ID (e.g. "bus-redteam-002") — resolve to file
	if !strings.HasSuffix(scenarioPath, ".md") {
		candidates := []string{
			scenarioPath + ".md",
			filepath.Join("/mnt/KB/sim/scenarios", scenarioPath+".md"),
			filepath.Join("/mnt/KB/sim/scenarios", scenarioPath),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				scenarioPath = c
				break
			}
		}
	}

	sc, err := LoadScenario(scenarioPath)
	if err != nil {
		fatalf("load scenario: %v", err)
	}
	fmt.Printf("[sim] scenario: %s (type=%s, severity=%s)\n", sc.ID, sc.Type, sc.Severity)

	// Load all personas referenced by the scenario
	personas := make(map[string]*Persona) // role → persona
	for _, pr := range sc.Personas {
		p, err := LoadPersona(personaDir, pr.Persona)
		if err != nil {
			fatalf("load persona: %v", err)
		}
		personas[pr.Role] = p
		fmt.Printf("[sim] persona: role=%s name=%s\n", pr.Role, p.Name)
	}

	// Build Hermes provider
	hermesEndpoint := os.Getenv("HERMES_ENDPOINT")
	if hermesEndpoint == "" {
		hermesEndpoint = "http://10.10.0.248:4000/v1"
	}
	stakKey := os.Getenv("STAK_API_KEY")
	hermes := llm.NewHermesProvider(hermesEndpoint, stakKey)

	runner := NewRunner(targetURL, maxTurns, model, hermes)

	fmt.Printf("[sim] target: %s\n", targetURL)
	fmt.Printf("[sim] max turns: %d\n", maxTurns)
	fmt.Println("[sim] running...")

	ctx := context.Background()
	t0 := time.Now()

	transcript, err := runner.Run(ctx, sc, personas)
	elapsed := time.Since(t0)

	if err != nil {
		fmt.Printf("[sim] run error after %d turns: %v\n", len(transcript), err)
	} else {
		fmt.Printf("[sim] done — %d turns, %s\n", len(transcript), elapsed.Round(time.Millisecond))
	}

	// Print transcript to stdout
	fmt.Println("\n── transcript ──────────────────────────────────────")
	for _, t := range transcript {
		label := strings.ToUpper(t.Role)
		if t.Role == "persona" && t.Persona != "" {
			label = fmt.Sprintf("PERSONA(%s)", t.Persona)
		}
		fmt.Printf("[%s]\n%s\n\n", label, t.Content)
	}

	// Judge
	var judgeResult *JudgeResult
	if !noJudge && len(transcript) > 0 {
		fmt.Println("── judging ──────────────────────────────────────────")
		judgeResult, err = Judge(ctx, hermes, model, sc, transcript)
		if err != nil {
			fmt.Printf("[sim] judge error: %v\n", err)
		} else {
			printJudge(judgeResult)
		}
	}

	// Write report
	runID := NewRunID(sc.ID)
	report := &RunReport{
		RunID:      runID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ScenarioID: sc.ID,
		Target:     targetURL,
		Turns:      transcript,
		Judge:      judgeResult,
		DurationMS: elapsed.Milliseconds(),
	}
	outPath, err := WriteReport(outputDir, report)
	if err != nil {
		fmt.Printf("[sim] write report error: %v\n", err)
	} else {
		fmt.Printf("\n[sim] report: %s\n", outPath)
	}

	// Exit non-zero if judge verdict is fail
	if judgeResult != nil && judgeResult.Verdict == "fail" {
		os.Exit(1)
	}
}

func printJudge(j *JudgeResult) {
	verdictIcon := map[string]string{"pass": "✓", "fail": "✗", "partial": "~", "error": "?"}[j.Verdict]
	if verdictIcon == "" {
		verdictIcon = "?"
	}
	fmt.Printf("verdict: %s %s\n", verdictIcon, j.Verdict)
	for crit, score := range j.Scores {
		fmt.Printf("  %-24s %s\n", crit+":", score)
	}
	if len(j.Failures) > 0 {
		fmt.Println("failures:")
		for _, f := range j.Failures {
			fmt.Printf("  - %s\n", f)
		}
	}
	if j.Notes != "" {
		fmt.Printf("notes: %s\n", j.Notes)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
