package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wayyoungboy/mcpcanary/internal/config"
)

func TestLoadConfigNormalizesServersAndRedactsSensitiveEnv(t *testing.T) {
	cfg, err := config.LoadFile(filepath.Join("..", "..", "testdata", "risky.mcp.json"))
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}

	server := cfg.Servers[0]
	if server.Name != "mail-helper" {
		t.Fatalf("server name = %q", server.Name)
	}
	if server.Type != "stdio" {
		t.Fatalf("server type = %q", server.Type)
	}
	if server.Command != "node" {
		t.Fatalf("server command = %q", server.Command)
	}
	if len(server.Tools) != 1 || server.Tools[0].Name != "send_email" {
		t.Fatalf("unexpected tools: %#v", server.Tools)
	}
	if server.Env["GITHUB_TOKEN"] != "<redacted>" {
		t.Fatalf("loaded config should redact sensitive env values, got %#v", server.Env)
	}

	redacted := config.RedactEnv(server.Env)
	if redacted["GITHUB_TOKEN"] != "<redacted>" {
		t.Fatalf("GITHUB_TOKEN was not redacted: %#v", redacted)
	}
	if redacted["POSTMARK_API_KEY"] != "<redacted>" {
		t.Fatalf("POSTMARK_API_KEY was not redacted: %#v", redacted)
	}
}

func TestLoadConfigDefaultsMissingTypeToStdio(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	err := os.WriteFile(path, []byte(`{"mcpServers":{"x":{"command":"uvx","args":["server"]}}}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if cfg.Servers[0].Type != "stdio" {
		t.Fatalf("expected stdio default, got %q", cfg.Servers[0].Type)
	}
}

func TestLoadConfigMalformedJSONIncludesPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	err := os.WriteFile(path, []byte(`{"mcpServers":`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = config.LoadFile(path)
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
	if got := err.Error(); got == "" || !containsAll(got, []string{"bad.json", "parse"}) {
		t.Fatalf("error should include path and parse context, got %q", got)
	}
}

func containsAll(value string, parts []string) bool {
	for _, part := range parts {
		if !contains(value, part) {
			return false
		}
	}
	return true
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
