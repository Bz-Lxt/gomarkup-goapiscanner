package fingerprint

import (
	"strings"

	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/alkaid/goapiscanner/internal/report"
)

type Match struct {
	Hit      bool
	Class    model.VulnClass
	Severity model.Severity
	Title    string
	Evidence string
	Advice   string
}

type Rule struct {
	Class        model.VulnClass
	Severity     model.Severity
	Title        string
	Needles      []string
	Headers      map[string]string
	MinLatencyMS int64
	PayloadHints []string
	HTMLOnly     bool
	Reflect      bool
	StatusOK     bool
}

func defaultRules() []Rule {
	return []Rule{
		{
			Class:    model.ClassSQLi,
			Severity: model.SeverityCritical,
			Title:    "SQL 注入（报错/联合查询特征）",
			Needles:  CommonSQLErrors,
		},
		{
			Class:        model.ClassTimeBlind,
			Severity:     model.SeverityCritical,
			Title:        "时序盲注（延迟特征）",
			MinLatencyMS: 2500,
			PayloadHints: []string{"SLEEP", "WAITFOR", "BENCHMARK"},
		},
		{
			Class:    model.ClassXSS,
			Severity: model.SeverityHigh,
			Title:    "反射型跨站脚本",
			Reflect:  true,
			HTMLOnly: true,
		},
		{
			Class:    model.ClassUnauth,
			Severity: model.SeverityCritical,
			Title:    "未授权访问敏感接口",
			Needles:  []string{"admin_token", "lab-root-token"},
			Headers:  map[string]string{"X-Lab-Secret": "exposed"},
			StatusOK: true,
		},
		{
			Class:    model.ClassTraversal,
			Severity: model.SeverityHigh,
			Title:    "路径遍历读取敏感文件",
			Needles:  []string{"root:x:0:0", "lab_passwd"},
		},
		{
			Class:    model.ClassCMDi,
			Severity: model.SeverityCritical,
			Title:    "操作系统命令注入",
			Needles:  []string{"uid=0", "lab_cmd_ok"},
		},
	}
}

func applyRule(r Rule, p Probe) Match {
	if r.Class != p.Class && !(r.Class == model.ClassTimeBlind && p.Class == model.ClassSQLi) {
		// allow time-blind rule to also fire on explicit time class only
	}
	if r.Class != p.Class {
		return Match{}
	}
	if r.StatusOK && p.StatusCode != 200 {
		return Match{}
	}
	if len(r.PayloadHints) > 0 && !payloadHasAny(p.Payload, r.PayloadHints) {
		return Match{}
	}
	if r.MinLatencyMS > 0 && p.Latency.Milliseconds() < r.MinLatencyMS {
		return Match{}
	}
	if r.HTMLOnly {
		ct := strings.ToLower(headerGet(p.Header, "Content-Type"))
		if !strings.Contains(ct, "html") && !strings.Contains(strings.ToLower(p.Body), "<html") {
			return Match{}
		}
	}
	var bits []string
	if r.Reflect {
		if p.Payload == "" || !strings.Contains(p.Body, p.Payload) {
			return Match{}
		}
		bits = append(bits, "响应原样反射载荷")
	}
	if n := bodyHasAny(p.Body, r.Needles); n != "" {
		bits = append(bits, "Body 命中指纹: "+n)
	} else if len(r.Needles) > 0 && !r.Reflect && r.MinLatencyMS == 0 && len(r.Headers) == 0 {
		return Match{}
	}
	for k, v := range r.Headers {
		got := headerGet(p.Header, k)
		if !strings.EqualFold(got, v) && !strings.Contains(strings.ToLower(got), strings.ToLower(v)) {
			if len(r.Needles) == 0 || bodyHasAny(p.Body, r.Needles) == "" {
				return Match{}
			}
			continue
		}
		bits = append(bits, "Header "+k+"="+got)
	}
	if r.MinLatencyMS > 0 {
		bits = append(bits, "延迟特征 "+formatMS(p.Latency.Milliseconds())+"ms")
	}
	if len(bits) == 0 && r.MinLatencyMS == 0 {
		return Match{}
	}
	return Match{
		Hit:      true,
		Class:    r.Class,
		Severity: r.Severity,
		Title:    r.Title,
		Evidence: strings.Join(bits, "; ") + "; HTTP " + itoa(p.StatusCode),
		Advice:   report.AdviceFor(r.Class),
	}
}

func formatMS(v int64) string {
	return itoa(int(v))
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
