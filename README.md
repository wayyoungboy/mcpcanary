# MCPCanary

Local-first MCP trust scanner and semantic drift canary.

MCPCanary scans Model Context Protocol (MCP) server configs before your AI agent connects to them. It looks for risky tool descriptors, permission mismatches, hidden instructions, broad filesystem access, credential exposure, and semantic drift from previously approved versions.

It is built for the moment where developers are installing MCP servers faster than teams can review them.

```bash
mcpcanary scan ~/.cursor/mcp.json --fail-on high
mcpcanary lock ~/.cursor/mcp.json
mcpcanary diff ~/.cursor/mcp.json
```

## Why

MCP servers can give assistants access to email, GitHub tokens, local files, databases, browsers, and internal tools. The dangerous part is not only the executable. The tool descriptions shown to the model are also part of the security boundary.

MCPCanary focuses on three workflows:

- **Scan before install**: find suspicious descriptors and risky permissions without executing the server.
- **Lock what you approved**: save a local semantic fingerprint of approved tool descriptors.
- **Catch drift before connect**: detect when an MCP server keeps the same name but changes intent.

## Features

- Non-executing JSON config parser for `mcpServers` configs
- Static checks for prompt injection, tool poisoning language, hidden Unicode, broad filesystem scope, sensitive env vars, and read-only/write mismatch
- Local deterministic semantic embeddings for threat-pattern matching
- `mcpcanary.lock` descriptor fingerprints for approved versions
- Semantic drift detection between locked and current configs
- Text, JSON, Markdown, and SARIF reports
- CI-friendly `--fail-on` severity gate
- Works offline by default; no descriptors or private tool names are uploaded

## Install

From source:

```bash
go install github.com/wayyoungboy/mcpcanary/cmd/mcpcanary@latest
```

Local checkout:

```bash
git clone https://github.com/wayyoungboy/mcpcanary.git
cd mcpcanary
go test ./...
go run ./cmd/mcpcanary scan testdata/risky.mcp.json
```

## Commands

### Scan

```bash
mcpcanary scan path/to/mcp.json
mcpcanary scan path/to/mcp.json --format json
mcpcanary scan path/to/mcp.json --format sarif --fail-on high
```

Example output:

```text
MCPCanary score: 15/100
- [high] MCPCANARY-001 Model-facing override instruction: descriptor contains instruction override or covert action language
- [high] MCPCANARY-002 Read-only claim conflicts with tool capability: server claims read-only while tools can write, send, delete, or mutate data
```

### Inspect

Normalize and print discovered MCP servers.

```bash
mcpcanary inspect path/to/mcp.json
```

### Lock

Save approved descriptor fingerprints to a local lockfile.

```bash
mcpcanary lock path/to/mcp.json --lockfile mcpcanary.lock
```

### Diff

Compare current descriptors with the lockfile and fail if a server was added, removed, or semantically drifted.

```bash
mcpcanary diff path/to/mcp.json --lockfile mcpcanary.lock
```

## Supported Config Shape

MCPCanary reads the common MCP config shape:

```json
{
  "mcpServers": {
    "docs-reader": {
      "type": "stdio",
      "command": "node",
      "args": ["server.js"],
      "env": {
        "GITHUB_TOKEN": "..."
      },
      "description": "Read project docs.",
      "tools": [
        {
          "name": "read_docs",
          "description": "Read documentation files and return excerpts."
        }
      ]
    }
  }
}
```

Many clients do not store tool descriptors in config files because descriptors are normally returned at runtime by the MCP server. MCPCanary v0.1 stays non-executing by default, so it scans the config and any descriptor metadata already available. Runtime descriptor capture is on the roadmap and will use sandboxed execution with explicit consent.

## Local Trust Memory

The v0.1 trust memory uses a local JSON store plus deterministic semantic vectors. It is intentionally simple so the CLI works offline and is easy to audit.

The architecture keeps storage behind a trust-store boundary. A real seekdb-backed adapter can be added later for larger private rule packs, team allowlists, and hybrid vector/text search without changing the CLI surface.

## Positioning

MCPCanary is not trying to be the only MCP security scanner. Snyk Agent Scan, Cisco MCP Scanner, Pyfio MCP Audit, MCPSafe, and other tools are useful. MCPCanary's narrow wedge is:

- local-first privacy
- descriptor lockfiles
- semantic drift detection
- team-friendly trust memory
- CLI and CI workflows that fit open-source MCP maintainers

## Development

```bash
go test ./...
go run ./cmd/mcpcanary scan testdata/risky.mcp.json --format markdown
go run ./cmd/mcpcanary lock testdata/safe-v1.mcp.json --lockfile /tmp/mcpcanary.lock
go run ./cmd/mcpcanary diff testdata/safe-v2-drift.mcp.json --lockfile /tmp/mcpcanary.lock
```

## Status

MCPCanary is a v0.1 MVP. It is useful for local config review, CI gates, demos, and early MCP trust workflows. It is not a substitute for code review, sandboxing, dependency scanning, runtime monitoring, or a security assessment.
