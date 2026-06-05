package model

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

const SchemaVersion = "v1"

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Config struct {
	Path    string   `json:"path"`
	Servers []Server `json:"servers"`
}

type Server struct {
	Name        string            `json:"name"`
	Type        string            `json:"type,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Package     string            `json:"package,omitempty"`
	Source      string            `json:"source,omitempty"`
	Description string            `json:"description,omitempty"`
	Tools       []Tool            `json:"tools,omitempty"`
	ConfigPath  string            `json:"config_path,omitempty"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

type Finding struct {
	ID             string   `json:"id"`
	Severity       Severity `json:"severity"`
	Title          string   `json:"title"`
	ServerName     string   `json:"server_name,omitempty"`
	Evidence       string   `json:"evidence"`
	Recommendation string   `json:"recommendation"`
}

type Report struct {
	SchemaVersion string    `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	ConfigPath    string    `json:"config_path,omitempty"`
	Servers       []Server  `json:"servers"`
	Findings      []Finding `json:"findings"`
	Score         int       `json:"score"`
}

type ThreatMatch struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

func NewReport(configPath string, servers []Server, findings []Finding, score int) Report {
	return Report{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   time.Now().UTC(),
		ConfigPath:    configPath,
		Servers:       servers,
		Findings:      findings,
		Score:         score,
	}
}

func (s Server) DescriptorText() string {
	var parts []string
	parts = append(parts, s.Name, s.Type, s.Command, strings.Join(s.Args, " "), s.URL, s.Package, s.Source, s.Description)
	for _, tool := range s.Tools {
		parts = append(parts, tool.Name, tool.Description)
	}
	return strings.Join(parts, "\n")
}

func (s Server) Fingerprint() string {
	envKeys := make([]string, 0, len(s.Env))
	for key := range s.Env {
		envKeys = append(envKeys, key)
	}
	sort.Strings(envKeys)
	envPart := strings.Join(envKeys, ",")
	sum := sha256.Sum256([]byte(s.DescriptorText() + "\n" + envPart))
	return hex.EncodeToString(sum[:])
}

func SeverityRank(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}
