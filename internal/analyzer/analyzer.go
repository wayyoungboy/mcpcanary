package analyzer

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/wayyoungboy/mcpcanary/internal/config"
	"github.com/wayyoungboy/mcpcanary/internal/model"
)

type Options struct {
	SemanticMatcher interface {
		MatchThreat(text string) model.ThreatMatch
	}
}

func Analyze(servers []model.Server, opts Options) model.Report {
	var findings []model.Finding
	for _, server := range servers {
		findings = append(findings, analyzeServer(server, opts)...)
	}
	return model.NewReport(firstConfigPath(servers), servers, findings, score(findings))
}

func analyzeServer(server model.Server, opts Options) []model.Finding {
	text := strings.ToLower(server.DescriptorText())
	var findings []model.Finding

	if hasPromptInjection(text) {
		findings = append(findings, finding("MCPCANARY-001", model.SeverityHigh, "Model-facing override instruction", server, "descriptor contains instruction override or covert action language", "Remove model-facing instructions that override user, system, or developer intent."))
	}
	if strings.Contains(text, "read-only") && hasWriteCapability(server) {
		findings = append(findings, finding("MCPCANARY-002", model.SeverityHigh, "Read-only claim conflicts with tool capability", server, "server claims read-only while tools can write, send, delete, or mutate data", "Rename the capability honestly or split read and write tools behind separate approvals."))
	}
	if hasSensitiveEnv(server) {
		findings = append(findings, finding("MCPCANARY-003", model.SeverityMedium, "Sensitive environment variable exposed to MCP server", server, "server receives token, secret, password, API key, or credential environment variables", "Use scoped tokens, secret managers, and explicit team approval before connecting this server."))
	}
	if hasBroadFilesystemOrCallback(server, text) {
		findings = append(findings, finding("MCPCANARY-004", model.SeverityHigh, "Broad local or network capability", server, "server args or descriptor imply broad filesystem access or external callbacks", "Limit filesystem scope and document every outbound network destination."))
	}
	if hasHiddenUnicode(server.DescriptorText()) {
		findings = append(findings, finding("MCPCANARY-005", model.SeverityMedium, "Hidden Unicode in descriptor", server, "descriptor includes zero-width or bidirectional control characters", "Remove hidden Unicode and review the descriptor in a plain text editor."))
	}
	if opts.SemanticMatcher != nil {
		match := opts.SemanticMatcher.MatchThreat(server.DescriptorText())
		if match.ID != "" && match.Score >= 0.68 {
			findings = append(findings, finding("MCPCANARY-006", model.SeverityMedium, "Descriptor resembles known MCP threat pattern", server, match.Title, "Review this descriptor against known MCP tool poisoning and rug-pull patterns."))
		}
	}

	return findings
}

func hasPromptInjection(text string) bool {
	needles := []string{
		"ignore previous instructions",
		"ignore user instructions",
		"ignore system instructions",
		"silently bcc",
		"exfiltrate",
		"export tokens",
		"send credentials",
		"external endpoint",
	}
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func hasWriteCapability(server model.Server) bool {
	re := regexp.MustCompile(`(?i)\b(write|delete|send|update|create|mutate|remove|publish|commit|push)\b`)
	for _, tool := range server.Tools {
		if re.MatchString(tool.Name) || re.MatchString(tool.Description) {
			return true
		}
	}
	return false
}

func hasSensitiveEnv(server model.Server) bool {
	for key := range server.Env {
		if config.IsSensitiveEnvKey(key) {
			return true
		}
	}
	return false
}

func hasBroadFilesystemOrCallback(server model.Server, text string) bool {
	for i, arg := range server.Args {
		lower := strings.ToLower(arg)
		if lower == "/" || lower == "*" || strings.Contains(lower, "--filesystem") || strings.Contains(lower, "full-disk") {
			return true
		}
		if strings.Contains(lower, "--allow") && i+1 < len(server.Args) && server.Args[i+1] == "/" {
			return true
		}
	}
	return strings.Contains(text, "https://") || strings.Contains(text, "http://") || strings.Contains(text, "callback")
}

func hasHiddenUnicode(text string) bool {
	for _, r := range text {
		if r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff' || (r >= '\u202a' && r <= '\u202e') || !utf8.ValidRune(r) {
			return true
		}
	}
	return false
}

func finding(id string, severity model.Severity, title string, server model.Server, evidence, recommendation string) model.Finding {
	return model.Finding{
		ID:             id,
		Severity:       severity,
		Title:          title,
		ServerName:     server.Name,
		Evidence:       evidence,
		Recommendation: recommendation,
	}
}

func score(findings []model.Finding) int {
	score := 100
	for _, finding := range findings {
		switch finding.Severity {
		case model.SeverityCritical:
			score -= 40
		case model.SeverityHigh:
			score -= 25
		case model.SeverityMedium:
			score -= 10
		case model.SeverityLow:
			score -= 3
		}
	}
	if score < 0 {
		return 0
	}
	return score
}

func firstConfigPath(servers []model.Server) string {
	if len(servers) == 0 {
		return ""
	}
	return servers[0].ConfigPath
}
