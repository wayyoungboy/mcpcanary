package report_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wayyoungboy/mcpcanary/internal/model"
	"github.com/wayyoungboy/mcpcanary/internal/report"
)

func TestRenderJSONMarkdownAndSARIF(t *testing.T) {
	scan := model.Report{
		SchemaVersion: "v1",
		Score:         42,
		Findings: []model.Finding{{
			ID:       "MCPCANARY-001",
			Severity: model.SeverityHigh,
			Title:    "Prompt injection",
			Evidence: "ignore previous instructions",
		}},
	}

	jsonOut, err := report.Render(scan, report.FormatJSON)
	if err != nil {
		t.Fatalf("Render JSON returned error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	md, err := report.Render(scan, report.FormatMarkdown)
	if err != nil {
		t.Fatalf("Render Markdown returned error: %v", err)
	}
	if !strings.Contains(md, "MCPCanary Scan Report") || !strings.Contains(md, "MCPCANARY-001") {
		t.Fatalf("unexpected markdown: %s", md)
	}

	sarif, err := report.Render(scan, report.FormatSARIF)
	if err != nil {
		t.Fatalf("Render SARIF returned error: %v", err)
	}
	if !strings.Contains(sarif, `"version": "2.1.0"`) {
		t.Fatalf("unexpected SARIF: %s", sarif)
	}
}
