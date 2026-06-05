package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wayyoungboy/mcpcanary/internal/model"
)

type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatSARIF    Format = "sarif"
)

func Render(scan model.Report, format Format) (string, error) {
	switch format {
	case "", FormatText:
		return renderText(scan), nil
	case FormatJSON:
		data, err := json.MarshalIndent(scan, "", "  ")
		return string(data), err
	case FormatMarkdown:
		return renderMarkdown(scan), nil
	case FormatSARIF:
		return renderSARIF(scan)
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}

func renderText(scan model.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "MCPCanary score: %d/100\n", scan.Score)
	if len(scan.Findings) == 0 {
		b.WriteString("No findings.\n")
		return b.String()
	}
	for _, finding := range scan.Findings {
		fmt.Fprintf(&b, "- [%s] %s %s: %s\n", finding.Severity, finding.ID, finding.Title, finding.Evidence)
	}
	return b.String()
}

func renderMarkdown(scan model.Report) string {
	var b strings.Builder
	b.WriteString("# MCPCanary Scan Report\n\n")
	fmt.Fprintf(&b, "**Score:** %d/100\n\n", scan.Score)
	if len(scan.Findings) == 0 {
		b.WriteString("No findings.\n")
		return b.String()
	}
	b.WriteString("| Severity | ID | Server | Finding | Evidence |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, finding := range scan.Findings {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", finding.Severity, finding.ID, finding.ServerName, finding.Title, finding.Evidence)
	}
	return b.String()
}

func renderSARIF(scan model.Report) (string, error) {
	type sarifResult struct {
		RuleID  string `json:"ruleId"`
		Level   string `json:"level"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	}
	doc := map[string]interface{}{
		"version": "2.1.0",
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"runs": []map[string]interface{}{{
			"tool": map[string]interface{}{
				"driver": map[string]interface{}{
					"name": "MCPCanary",
				},
			},
			"results": []sarifResult{},
		}},
	}
	results := make([]sarifResult, 0, len(scan.Findings))
	for _, finding := range scan.Findings {
		var result sarifResult
		result.RuleID = finding.ID
		result.Level = sarifLevel(finding.Severity)
		result.Message.Text = finding.Title + ": " + finding.Evidence
		results = append(results, result)
	}
	doc["runs"].([]map[string]interface{})[0]["results"] = results
	data, err := json.MarshalIndent(doc, "", "  ")
	return string(data), err
}

func sarifLevel(severity model.Severity) string {
	switch severity {
	case model.SeverityCritical, model.SeverityHigh:
		return "error"
	case model.SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}
