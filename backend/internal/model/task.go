package model

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID           string     `json:"id"`
	BaseURL      string     `json:"base_url"`
	Status       TaskStatus `json:"status"`
	Concurrency  int        `json:"concurrency"`
	TimeoutMS    int        `json:"timeout_ms"`
	Authorized   bool       `json:"authorized"`
	SwaggerName  string     `json:"swagger_name"`
	Total        int        `json:"total"`
	Sent         int        `json:"sent"`
	Hits         int        `json:"hits"`
	Error        string     `json:"error"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
	Critical     int        `json:"critical"`
	High         int        `json:"high"`
	Medium       int        `json:"medium"`
	Low          int        `json:"low"`
	Info         int        `json:"info"`
}

type Finding struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	Endpoint   string    `json:"endpoint"`
	Method     string    `json:"method"`
	Class      VulnClass `json:"class"`
	Severity   Severity  `json:"severity"`
	Title      string    `json:"title"`
	Evidence   string    `json:"evidence"`
	Payload    string    `json:"payload"`
	ParamName  string    `json:"param_name"`
	StatusCode int       `json:"status_code"`
	LatencyMS  int64     `json:"latency_ms"`
	Advice     string    `json:"advice"`
	CreatedAt  string    `json:"created_at"`
}

type DefectNode struct {
	Key      string       `json:"key"`
	Label    string       `json:"label"`
	Method   string       `json:"method,omitempty"`
	Path     string       `json:"path,omitempty"`
	Severity Severity     `json:"severity,omitempty"`
	Count    int          `json:"count"`
	Finding  *Finding     `json:"finding,omitempty"`
	Children []DefectNode `json:"children"`
}

type SeverityStats struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}
