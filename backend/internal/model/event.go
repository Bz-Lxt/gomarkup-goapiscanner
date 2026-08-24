package model

type EventType string

const (
	EventLog      EventType = "log"
	EventProgress EventType = "progress"
	EventFinding  EventType = "finding"
	EventDone     EventType = "done"
)

type StreamEvent struct {
	Type     EventType `json:"type"`
	TS       string    `json:"ts"`
	Level    string    `json:"level,omitempty"`
	Message  string    `json:"message,omitempty"`
	Sent     int       `json:"sent,omitempty"`
	Total    int       `json:"total,omitempty"`
	Hits     int       `json:"hits,omitempty"`
	Finding  *Finding  `json:"finding,omitempty"`
	Status   string    `json:"status,omitempty"`
	TaskID   string    `json:"task_id"`
}

type CreateScanRequest struct {
	BaseURL     string `json:"base_url"`
	Concurrency int    `json:"concurrency"`
	TimeoutMS   int    `json:"timeout_ms"`
	Authorized  bool   `json:"authorized"`
}

type ReportPreview struct {
	Task     Task      `json:"task"`
	Findings []Finding `json:"findings"`
	Tree     []DefectNode `json:"tree"`
	Stats    SeverityStats `json:"stats"`
	Advice   []string  `json:"advice"`
	GeneratedAt string `json:"generated_at"`
}
