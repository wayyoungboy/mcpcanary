# Release Notes

## v0.1.0

Initial open-source MVP.

### Included

- `scan` command for local MCP config review
- `inspect` command for normalized server metadata
- `lock` command for approved descriptor fingerprints
- `diff` command for semantic drift detection
- Text, JSON, Markdown, and SARIF output
- CI-friendly `--fail-on` severity gate
- Offline deterministic semantic vectors
- Local trust store with seeded MCP threat patterns
- GitHub Actions test workflow

### Not Included Yet

- Sandboxed runtime descriptor capture
- Real seekdb backend adapter
- Hosted trust feed
- Team dashboard
- Package reputation crawling
- Homebrew tap

### Verification

```bash
go test ./...
go run ./cmd/mcpcanary scan testdata/risky.mcp.json --format json
go run ./cmd/mcpcanary lock testdata/safe-v1.mcp.json --lockfile /tmp/mcpcanary.lock
go run ./cmd/mcpcanary diff testdata/safe-v2-drift.mcp.json --lockfile /tmp/mcpcanary.lock
```
