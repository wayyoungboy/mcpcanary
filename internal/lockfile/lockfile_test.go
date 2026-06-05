package lockfile_test

import (
	"path/filepath"
	"testing"

	"github.com/wayyoungboy/mcpcanary/internal/config"
	"github.com/wayyoungboy/mcpcanary/internal/lockfile"
)

func TestLockfileDetectsSemanticDrift(t *testing.T) {
	v1, err := config.LoadFile(filepath.Join("..", "..", "testdata", "safe-v1.mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	v2, err := config.LoadFile(filepath.Join("..", "..", "testdata", "safe-v2-drift.mcp.json"))
	if err != nil {
		t.Fatal(err)
	}

	lock := lockfile.FromServers(v1.Servers, 100)
	diff := lockfile.Compare(lock, v2.Servers, lockfile.CompareOptions{DriftThreshold: 0.18})

	if len(diff.Changes) == 0 {
		t.Fatal("expected drift change")
	}
	if diff.Changes[0].Kind != lockfile.ChangeDrifted {
		t.Fatalf("expected drifted change, got %#v", diff.Changes[0])
	}
}

func TestLockfileRoundTrip(t *testing.T) {
	cfg, err := config.LoadFile(filepath.Join("..", "..", "testdata", "safe-v1.mcp.json"))
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "mcpcanary.lock")
	original := lockfile.FromServers(cfg.Servers, 100)
	if err := lockfile.Write(path, original); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	loaded, err := lockfile.Read(path)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if len(loaded.Servers) != 1 || loaded.Servers[0].Name != "docs-reader" {
		t.Fatalf("unexpected loaded lockfile: %#v", loaded)
	}
}
