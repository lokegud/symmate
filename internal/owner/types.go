package owner

// item is the per-conversation row returned by listConversations.
type item struct {
	ID            string  `json:"id"`
	VisitorHandle *string `json:"visitor_handle,omitempty"`
	Surface       string  `json:"surface"`
	StartedAt     int64   `json:"started_at"`
	EndedAt       *int64  `json:"ended_at,omitempty"`
	Outcome       *string `json:"outcome,omitempty"`
}

// msgItem is the per-message row returned by getConversation.
type msgItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	TS      int64  `json:"ts"`
}
