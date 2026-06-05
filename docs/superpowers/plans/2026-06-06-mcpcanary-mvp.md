# MCPCanary MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a releaseable open-source MCPCanary v0.1 CLI that scans MCP configs without executing untrusted servers, reports security risks, stores local semantic trust fingerprints, creates lockfiles, and detects descriptor drift.

**Architecture:** The CLI is a small Go module with focused internal packages: config discovery/parsing, analyzer rules, deterministic semantic vectors, trust store, lockfile diff, report rendering, and CLI commands. The MVP uses a local JSON trust store with deterministic embeddings so it works offline; the storage interface is named for the seekdb trust-memory concept and can later gain a real seekdb adapter without changing command behavior.

**Tech Stack:** Go 1.22+, standard library only for the core CLI, `go test ./...` for verification, GitHub Actions for CI, Homebrew/go-install friendly release metadata.

---

## File Structure

- `go.mod`: module definition for `github.com/wayyoungboy/mcpcanary`.
- `cmd/mcpcanary/main.go`: entrypoint delegating to `internal/cli`.
- `internal/cli/cli.go`: command parsing, exit codes, command orchestration.
- `internal/config/config.go`: MCP config structs, JSON parsing, server normalization, auto-discovery paths.
- `internal/model/model.go`: shared domain types for servers, findings, reports, trust records, lockfiles.
- `internal/analyzer/analyzer.go`: static rule checks and scoring.
- `internal/semantic/vector.go`: deterministic token-hash embeddings and cosine similarity.
- `internal/trust/store.go`: local semantic trust store for approved descriptors and known threat patterns.
- `internal/lockfile/lockfile.go`: lockfile creation and drift comparison.
- `internal/report/report.go`: text, JSON, Markdown, and SARIF output.
- `testdata/`: sample safe, risky, and drift MCP configs.
- `.github/workflows/ci.yml`: test workflow.
- `README.md`, `docs/THREAT_MODEL.md`, `docs/RELEASE.md`, `LICENSE`, `.gitignore`: release-ready project definition.

## Task 1: Project Skeleton and Config Parser

**Files:**
- Create: `go.mod`
- Create: `cmd/mcpcanary/main.go`
- Create: `internal/model/model.go`
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`
- Create: `testdata/risky.mcp.json`

- [ ] **Step 1: Write failing config parser tests**

Create `internal/config/config_test.go` with tests that parse a Cursor/Claude-style `mcpServers` map, redact sensitive env keys, and reject malformed JSON with a useful error.

- [ ] **Step 2: Run parser tests and confirm RED**

Run: `go test ./internal/config -run TestLoadConfig -v`

Expected: FAIL because `internal/config` does not exist yet.

- [ ] **Step 3: Implement minimal model and config parser**

Create the model structs and parser using `encoding/json`; normalize missing `type` to `stdio`; retain command, args, env, URL, and raw config path.

- [ ] **Step 4: Run parser tests and confirm GREEN**

Run: `go test ./internal/config -v`

Expected: PASS.

## Task 2: Analyzer Rules and Risk Scoring

**Files:**
- Create: `internal/analyzer/analyzer.go`
- Test: `internal/analyzer/analyzer_test.go`
- Modify: `internal/model/model.go`

- [ ] **Step 1: Write failing analyzer tests**

Cover hidden Unicode, suspicious descriptor instructions, read-only mismatch with write/delete/send tools, broad filesystem access, env/token access, shell execution, and risk score aggregation.

- [ ] **Step 2: Run analyzer tests and confirm RED**

Run: `go test ./internal/analyzer -v`

Expected: FAIL because analyzer functions are missing.

- [ ] **Step 3: Implement minimal analyzer**

Implement deterministic rules that return findings with ID, severity, title, evidence, and recommendation. Map severities to a 0-100 score where critical/high findings reduce score sharply.

- [ ] **Step 4: Run analyzer tests and confirm GREEN**

Run: `go test ./internal/analyzer -v`

Expected: PASS.

## Task 3: Semantic Trust Memory

**Files:**
- Create: `internal/semantic/vector.go`
- Test: `internal/semantic/vector_test.go`
- Create: `internal/trust/store.go`
- Test: `internal/trust/store_test.go`
- Modify: `internal/analyzer/analyzer.go`

- [ ] **Step 1: Write failing semantic tests**

Cover deterministic embeddings, cosine similarity, similar malicious descriptor detection, and false-positive memory persistence.

- [ ] **Step 2: Run semantic/trust tests and confirm RED**

Run: `go test ./internal/semantic ./internal/trust -v`

Expected: FAIL because packages are missing.

- [ ] **Step 3: Implement local vector store**

Use token-hash embeddings with fixed dimension 128. Store records in JSON under a caller-provided path. Seed known threat phrases in memory and use cosine similarity to emit semantic risk findings.

- [ ] **Step 4: Run tests and confirm GREEN**

Run: `go test ./internal/semantic ./internal/trust ./internal/analyzer -v`

Expected: PASS.

## Task 4: Lockfile and Drift Detection

**Files:**
- Create: `internal/lockfile/lockfile.go`
- Test: `internal/lockfile/lockfile_test.go`
- Create: `testdata/safe-v1.mcp.json`
- Create: `testdata/safe-v2-drift.mcp.json`

- [ ] **Step 1: Write failing lockfile tests**

Cover lock creation from scan report, descriptor hash changes, semantic drift threshold, added/removed servers, and stable same-version comparison.

- [ ] **Step 2: Run lockfile tests and confirm RED**

Run: `go test ./internal/lockfile -v`

Expected: FAIL because package is missing.

- [ ] **Step 3: Implement lockfile read/write/diff**

Write `mcpcanary.lock` as formatted JSON. Include schema version, server name, package source, command, descriptor hash, vector, approved risk score, and timestamp.

- [ ] **Step 4: Run lockfile tests and confirm GREEN**

Run: `go test ./internal/lockfile -v`

Expected: PASS.

## Task 5: Report Renderers and CLI Commands

**Files:**
- Create: `internal/report/report.go`
- Test: `internal/report/report_test.go`
- Create: `internal/cli/cli.go`
- Test: `internal/cli/cli_test.go`
- Modify: `cmd/mcpcanary/main.go`

- [ ] **Step 1: Write failing report/CLI tests**

Cover `scan`, `inspect`, `lock`, `diff`, `version`, text output, JSON output, Markdown output, SARIF output, and non-zero exit code when `--fail-on high` is exceeded.

- [ ] **Step 2: Run report/CLI tests and confirm RED**

Run: `go test ./internal/report ./internal/cli -v`

Expected: FAIL because packages are missing.

- [ ] **Step 3: Implement renderers and command runner**

Use standard-library flag parsing. `scan` reads one config or discovers known paths. `inspect` prints normalized servers. `lock` writes `mcpcanary.lock`. `diff` compares against a lockfile. Output is selected by `--format text|json|markdown|sarif`.

- [ ] **Step 4: Run report/CLI tests and confirm GREEN**

Run: `go test ./internal/report ./internal/cli -v`

Expected: PASS.

## Task 6: Open-Source Release Polish

**Files:**
- Create: `README.md`
- Create: `docs/THREAT_MODEL.md`
- Create: `docs/RELEASE.md`
- Create: `.github/workflows/ci.yml`
- Create: `.gitignore`
- Create: `LICENSE`
- Create: `examples/`

- [ ] **Step 1: Write docs smoke check**

Run `rg -n "MCPCanary|scan|lock|diff|local-first|MCP" README.md docs || true` after docs exist and confirm key terms are present.

- [ ] **Step 2: Add release docs and CI**

Document install commands, threat model, command examples, output samples, privacy guarantees, competitor-aware positioning, and contribution flow. Add GitHub Actions to run `go test ./...`.

- [ ] **Step 3: Run full verification**

Run:

```bash
gofmt -w .
go test ./...
go run ./cmd/mcpcanary scan testdata/risky.mcp.json --format json
go run ./cmd/mcpcanary lock testdata/safe-v1.mcp.json --lockfile /tmp/mcpcanary.lock
go run ./cmd/mcpcanary diff testdata/safe-v2-drift.mcp.json --lockfile /tmp/mcpcanary.lock
```

Expected: tests pass; scan emits findings; lock writes a file; diff reports semantic drift.

## Task 7: GitHub Publication

**Files:**
- Modify: git index only

- [ ] **Step 1: Commit releaseable MVP**

Run:

```bash
git status --short
git add .
git commit -m "feat: release MCPCanary MVP"
```

- [ ] **Step 2: Create GitHub repository**

Prefer GitHub connector if available. If `gh` works, run:

```bash
gh repo create wayyoungboy/mcpcanary --public --source=. --remote=origin --description "Local-first MCP trust scanner and semantic drift canary" --push
```

- [ ] **Step 3: Verify published repo**

Run:

```bash
git remote -v
git status --short
gh repo view wayyoungboy/mcpcanary --web=false
```

Expected: remote points to GitHub, working tree clean, repository metadata is visible.

---

## Self-Review

- Spec coverage: The plan covers local scanning, non-execution by default, static rules, semantic trust memory, lockfile, drift, CI, docs, and GitHub publication.
- Scope: Hosted trust feed, SSO, private registry, Homebrew tap, and full package reputation crawling are intentionally excluded from v0.1 because they are not required for an open-source releaseable MVP.
- Placeholder scan: No TBD/TODO/implement-later placeholders are present.
