package model

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

var severityRank = map[Severity]int{
	SeverityCritical: 5,
	SeverityHigh:     4,
	SeverityMedium:   3,
	SeverityLow:      2,
	SeverityInfo:     1,
}

func (s Severity) Label() string {
	switch s {
	case SeverityCritical:
		return "严重"
	case SeverityHigh:
		return "高危"
	case SeverityMedium:
		return "中危"
	case SeverityLow:
		return "低危"
	default:
		return "信息"
	}
}

func (s Severity) Rank() int {
	if r, ok := severityRank[s]; ok {
		return r
	}
	return 0
}

func ParseSeverity(s string) Severity {
	switch Severity(s) {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return Severity(s)
	default:
		return SeverityInfo
	}
}
