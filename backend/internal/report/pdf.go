package report

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/go-pdf/fpdf"
)

func RenderPDF(task model.Task, findings []model.Finding, fontPath string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 18)
	useCJK := false
	if fontPath != "" {
		if _, err := os.Stat(fontPath); err == nil {
			pdf.AddUTF8Font("cjk", "", fontPath)
			pdf.AddUTF8Font("cjk", "B", fontPath)
			useCJK = true
		}
	}
	setFont := func(style string, size float64) {
		if useCJK {
			pdf.SetFont("cjk", style, size)
			return
		}
		if style == "B" {
			pdf.SetFont("Helvetica", "B", size)
			return
		}
		pdf.SetFont("Helvetica", "", size)
	}
	tr := func(s string) string {
		if useCJK {
			return s
		}
		return asciiFallback(s)
	}

	pdf.AddPage()
	pdf.SetFillColor(11, 18, 32)
	pdf.Rect(0, 0, 210, 42, "F")
	pdf.SetTextColor(62, 224, 197)
	setFont("B", 18)
	pdf.SetXY(16, 12)
	pdf.Cell(0, 10, tr("GoAPIScanner 安全合规报告"))
	pdf.SetTextColor(200, 214, 229)
	setFont("", 11)
	pdf.SetXY(16, 24)
	pdf.Cell(0, 8, tr("任务 "+task.ID+"  ·  生成时间 "+task.UpdatedAt))

	pdf.SetTextColor(20, 24, 32)
	pdf.SetY(50)
	setFont("B", 13)
	pdf.Cell(0, 8, tr("一、扫描摘要"))
	pdf.Ln(10)
	setFont("", 11)
	lines := []string{
		fmt.Sprintf("目标: %s", task.BaseURL),
		fmt.Sprintf("状态: %s    请求: %d / %d    命中: %d", task.Status, task.Sent, task.Total, task.Hits),
		fmt.Sprintf("严重 %d  高危 %d  中危 %d  低危 %d  信息 %d", task.Critical, task.High, task.Medium, task.Low, task.Info),
	}
	for _, ln := range lines {
		pdf.MultiCell(0, 6, tr(ln), "", "", false)
	}

	pdf.Ln(4)
	setFont("B", 13)
	pdf.Cell(0, 8, tr("二、修复建议"))
	pdf.Ln(10)
	setFont("", 11)
	for i, a := range AdviceList(findings) {
		pdf.MultiCell(0, 6, tr(fmt.Sprintf("%d. %s", i+1, a)), "", "", false)
		pdf.Ln(1)
	}

	pdf.Ln(4)
	setFont("B", 13)
	pdf.Cell(0, 8, tr("三、漏洞明细"))
	pdf.Ln(10)
	if len(findings) == 0 {
		setFont("", 11)
		pdf.Cell(0, 6, tr("未发现指纹命中。"))
	}
	for i, f := range findings {
		setFont("B", 11)
		pdf.SetTextColor(severityRGB(f.Severity))
		pdf.MultiCell(0, 6, tr(fmt.Sprintf("%d. [%s] %s", i+1, f.Severity.Label(), f.Title)), "", "", false)
		pdf.SetTextColor(20, 24, 32)
		setFont("", 10)
		pdf.MultiCell(0, 5, tr(fmt.Sprintf("%s %s  参数 %s", f.Method, f.Endpoint, f.ParamName)), "", "", false)
		pdf.MultiCell(0, 5, tr("证据: "+f.Evidence), "", "", false)
		pdf.MultiCell(0, 5, tr("载荷: "+clip(f.Payload, 180)), "", "", false)
		pdf.MultiCell(0, 5, tr("建议: "+f.Advice), "", "", false)
		pdf.Ln(3)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func severityRGB(s model.Severity) (int, int, int) {
	switch s {
	case model.SeverityCritical:
		return 232, 72, 85
	case model.SeverityHigh:
		return 232, 140, 48
	case model.SeverityMedium:
		return 230, 196, 67
	default:
		return 90, 110, 130
	}
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func asciiFallback(s string) string {
	repl := map[string]string{
		"安全合规报告": "Security Compliance Report",
		"任务":     "Task",
		"生成时间":   "Generated",
		"一、扫描摘要": "1. Summary",
		"二、修复建议": "2. Remediation",
		"三、漏洞明细": "3. Findings",
		"目标":     "Target",
		"状态":     "Status",
		"请求":     "Requests",
		"命中":     "Hits",
		"严重":     "Critical",
		"高危":     "High",
		"中危":     "Medium",
		"低危":     "Low",
		"信息":     "Info",
		"参数":     "param",
		"证据":     "Evidence",
		"载荷":     "Payload",
		"建议":     "Advice",
		"未发现指纹命中。": "No fingerprint hit.",
	}
	out := s
	for k, v := range repl {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}
