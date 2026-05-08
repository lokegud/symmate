package visitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ForcefieldBaseMessage is the prefix sent when a message is blocked.
// If a report ID is available it is appended; callers should use BlockedMessage().
const ForcefieldBaseMessage = "Forcefield blocked: Sorry, we're being a little extra protective today. " +
	"You can either: retry your prompt, ask something else, or we can take a closer look."

// BlockedMessage returns the full message to send to a caller on a block.
// When reportID is non-empty it appends a reference the caller can use to retrieve the report.
func BlockedMessage(reportID string) string {
	if reportID == "" {
		return ForcefieldBaseMessage
	}
	return fmt.Sprintf("%s (ref: %s)", ForcefieldBaseMessage, reportID)
}

// Forcefield is a thin client to Poker's /diffuse endpoint.
// If Endpoint is empty or Poker is unreachable, Check returns allow=true (fail open).
type Forcefield struct {
	Endpoint string // e.g. "http://localhost:7437"
	client   *http.Client
}

type diffuseRequest struct {
	CallerID   string `json:"caller_id"`
	CallerType string `json:"caller_type"`
	Message    string `json:"message"`
	Context    string `json:"context"`
}

type diffuseResult struct {
	Allow       bool   `json:"allow"`
	ThreatClass string `json:"threat_class"`
	Reason      string `json:"reason"`
	AuditNote   string `json:"audit_note"`
	ReportID    string `json:"report_id"`
}

func NewForcefield(endpoint string) *Forcefield {
	return &Forcefield{
		Endpoint: endpoint,
		client:   &http.Client{Timeout: 3 * time.Second},
	}
}

// Check evaluates an inbound message. Returns (allow, reportID, auditNote).
// reportID is non-empty only when the message was blocked and a report was recorded.
// Fails open: if Poker is unreachable or endpoint is empty, allow = true.
func (f *Forcefield) Check(ctx context.Context, callerID, callerType, message, hostContext string) (allow bool, reportID string, auditNote string) {
	if f.Endpoint == "" {
		return true, "", ""
	}

	body, _ := json.Marshal(diffuseRequest{
		CallerID:   callerID,
		CallerType: callerType,
		Message:    message,
		Context:    hostContext,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", f.Endpoint+"/diffuse", bytes.NewReader(body))
	if err != nil {
		return true, "", ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		// Poker down — fail open, don't block legitimate users
		return true, "", ""
	}
	defer resp.Body.Close()

	var result diffuseResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, "", ""
	}

	return result.Allow, result.ReportID, result.AuditNote
}
