package eventbus

import "time"

// EventType is the canonical name of a domain event (e.g. "deal.closed").
type EventType string

const (
	// Sales
	DealClosed  EventType = "deal.closed"
	DealUpdated EventType = "deal.updated"

	// HR
	HireStarted   EventType = "hire.started"
	HireCompleted EventType = "hire.completed"

	// Security / Ops
	IncidentRaised   EventType = "incident.raised"
	IncidentResolved EventType = "incident.resolved"

	// Legal / Finance
	ContractSigned EventType = "contract.signed"
	InvoicePaid    EventType = "invoice.paid"

	// Data
	MetricRecorded EventType = "metric.recorded"
)

// Event is the envelope carried on the bus.
type Event struct {
	ID         string    `json:"id"`
	Type       EventType `json:"type"`
	Source     string    `json:"source"`      // agent or service that emitted the event
	OccurredAt time.Time `json:"occurred_at"` // wall-clock time at emission
	Payload    any       `json:"payload"`
}

// --- typed payloads ---

type DealClosedPayload struct {
	DealID      string  `json:"deal_id"`
	AccountName string  `json:"account_name"`
	ARR         float64 `json:"arr"`
	OwnerEmail  string  `json:"owner_email"`
	Won         bool    `json:"won"`
}

type DealUpdatedPayload struct {
	DealID   string `json:"deal_id"`
	OldStage string `json:"old_stage"`
	NewStage string `json:"new_stage"`
}

type HireStartedPayload struct {
	EmployeeID string `json:"employee_id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	StartDate  string `json:"start_date"`
}

type HireCompletedPayload struct {
	EmployeeID string `json:"employee_id"`
	Name       string `json:"name"`
}

type IncidentRaisedPayload struct {
	IncidentID string `json:"incident_id"`
	Title      string `json:"title"`
	Severity   string `json:"severity"` // low | medium | high | critical
	Source     string `json:"source"`
}

type IncidentResolvedPayload struct {
	IncidentID     string `json:"incident_id"`
	ResolutionNote string `json:"resolution_note"`
	DurationSecs   int64  `json:"duration_secs"`
}

type ContractSignedPayload struct {
	ContractID  string  `json:"contract_id"`
	AccountName string  `json:"account_name"`
	Value       float64 `json:"value"`
	SignedBy    string  `json:"signed_by"`
}

type InvoicePaidPayload struct {
	InvoiceID   string  `json:"invoice_id"`
	AccountName string  `json:"account_name"`
	Amount      float64 `json:"amount"`
}

type MetricRecordedPayload struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
	Tags  map[string]string `json:"tags,omitempty"`
}
