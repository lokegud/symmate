package marketing

// CampaignStatus represents the lifecycle state of a marketing campaign.
type CampaignStatus string

const (
	StatusDraft     CampaignStatus = "draft"
	StatusActive    CampaignStatus = "active"
	StatusPaused    CampaignStatus = "paused"
	StatusCompleted CampaignStatus = "completed"
)

// Campaign represents a single marketing campaign.
type Campaign struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Channel   string         `json:"channel"` // email | linkedin | events | paid_search
	Status    CampaignStatus `json:"status"`
	Budget    float64        `json:"budget"`
	Spend     float64        `json:"spend"`
	Leads     int            `json:"leads"`
	MQLs      int            `json:"mqls"`
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date"`
}

// ContentItem represents a piece of marketing content.
type ContentItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`   // blog | case_study | webinar | whitepaper
	Status      string `json:"status"` // draft | published
	PublishDate string `json:"publish_date"`
	Views       int    `json:"views"`
}

// AttributionReport summarises campaign impact for a period.
type AttributionReport struct {
	Period             string     `json:"period"`
	Campaigns          []Campaign `json:"campaigns"`
	TotalLeads         int        `json:"total_leads"`
	TotalMQLs          int        `json:"total_mqls"`
	PipelineInfluenced float64    `json:"pipeline_influenced"`
}
