package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wayyoungboy/mcpcanary/internal/cli"
)

func TestScanCommandOutputsFindingsAndHonorsFailOn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"scan", filepath.Join("..", "..", "testdata", "risky.mcp.json"), "--format", "json", "--fail-on", "high"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("expected non-zero exit for high findings")
	}
	if !strings.Contains(stdout.String(), "MCPCANARY-001") {
		t.Fatalf("expected finding in stdout, got %s", stdout.String())
	}
}

func TestLockAndDiffCommandsReportDrift(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "mcpcanary.lock")

	var lockOut, lockErr bytes.Buffer
	lockCode := cli.Run([]string{"lock", filepath.Join("..", "..", "testdata", "safe-v1.mcp.json"), "--lockfile", lockPath}, &lockOut, &lockErr)
	if lockCode != 0 {
		t.Fatalf("lock command failed: stdout=%s stderr=%s", lockOut.String(), lockErr.String())
	}

	var diffOut, diffErr bytes.Buffer
	diffCode := cli.Run([]string{"diff", filepath.Join("..", "..", "testdata", "safe-v2-drift.mcp.json"), "--lockfile", lockPath}, &diffOut, &diffErr)
	if diffCode == 0 {
		t.Fatal("expected diff command to return non-zero for drift")
	}
	if !strings.Contains(diffOut.String(), "drifted") {
		t.Fatalf("expected drift output, got stdout=%s stderr=%s", diffOut.String(), diffErr.String())
	}
}

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("version failed: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "MCPCanary") {
		t.Fatalf("unexpected version output: %s", stdout.String())
	}
}
