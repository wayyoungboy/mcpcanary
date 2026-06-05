package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wayyoungboy/mcpcanary/internal/model"
)

type rawConfig struct {
	MCPServers map[string]rawServer `json:"mcpServers"`
}

type rawServer struct {
	Type        string            `json:"type"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	URL         string            `json:"url"`
	Package     string            `json:"package"`
	Source      string            `json:"source"`
	Description string            `json:"description"`
	Tools       []model.Tool      `json:"tools"`
}

func LoadFile(path string) (model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return model.Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(raw.MCPServers) == 0 {
		return model.Config{Path: path}, nil
	}

	names := make([]string, 0, len(raw.MCPServers))
	for name := range raw.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)

	servers := make([]model.Server, 0, len(names))
	for _, name := range names {
		item := raw.MCPServers[name]
		serverType := item.Type
		if serverType == "" {
			serverType = "stdio"
		}
		servers = append(servers, model.Server{
			Name:        name,
			Type:        serverType,
			Command:     item.Command,
			Args:        item.Args,
			Env:         RedactEnv(item.Env),
			URL:         item.URL,
			Package:     item.Package,
			Source:      item.Source,
			Description: item.Description,
			Tools:       item.Tools,
			ConfigPath:  path,
		})
	}

	return model.Config{Path: path, Servers: servers}, nil
}

func RedactEnv(env map[string]string) map[string]string {
	redacted := make(map[string]string, len(env))
	for key, value := range env {
		if IsSensitiveEnvKey(key) {
			redacted[key] = "<redacted>"
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func IsSensitiveEnvKey(key string) bool {
	upper := strings.ToUpper(key)
	needles := []string{"TOKEN", "SECRET", "PASSWORD", "API_KEY", "ACCESS_KEY", "PRIVATE_KEY", "CREDENTIAL"}
	for _, needle := range needles {
		if strings.Contains(upper, needle) {
			return true
		}
	}
	return false
}

func Discover(home string) []string {
	candidates := []string{
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(home, ".vscode", "mcp.json"),
		filepath.Join(home, ".claude", "mcp.json"),
		filepath.Join(home, ".codex", "mcp.json"),
		filepath.Join(home, ".config", "claude", "mcp.json"),
		filepath.Join(home, ".config", "mcpcanary", "mcp.json"),
	}
	var found []string
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			found = append(found, candidate)
		}
	}
	return found
}
