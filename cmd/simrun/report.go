package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunReport is the full record of one sim run, written to the output dir.
type RunReport struct {
	RunID      string       `json:"run_id"`
	Timestamp  string       `json:"timestamp"`
	ScenarioID string       `json:"scenario_id"`
	Target     string       `json:"target"`
	Turns      []Turn       `json:"transcript"`
	Judge      *JudgeResult `json:"judge"`
	DurationMS int64        `json:"duration_ms"`
}

// WriteReport writes the run report as a JSONL file (one JSON object per line)
// to outputDir/<scenario_id>/<run_id>.jsonl
func WriteReport(outputDir string, report *RunReport) (string, error) {
	dir := filepath.Join(outputDir, report.ScenarioID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	filename := filepath.Join(dir, report.RunID+".jsonl")
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("create report file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(report); err != nil {
		return "", fmt.Errorf("encode report: %w", err)
	}

	return filename, nil
}

// NewRunID generates a run ID in the form YYYYMMDD-HHMMSS-<scenario_id>.
func NewRunID(scenarioID string) string {
	now := time.Now()
	ts := now.Format("20060102-150405")
	// shorten scenario id to first segment for readability
	short := scenarioID
	return fmt.Sprintf("%s-%s", ts, short)
}
