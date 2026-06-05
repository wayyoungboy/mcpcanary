# MCPCanary Threat Model

## Assets

- MCP server descriptors shown to an AI model
- Local MCP config files
- Developer credentials passed through environment variables
- Local filesystem paths exposed through stdio servers
- Team approval history stored in `mcpcanary.lock`
- Private tool names and internal workflow descriptions

## Threats In Scope

- Tool poisoning through model-facing descriptor instructions
- Tool shadowing and misleading tool names
- Read-only claims that hide write/send/delete capabilities
- Hidden Unicode or obfuscated descriptor text
- Broad filesystem access through args such as `/` or `--filesystem`
- Sensitive env var exposure such as tokens, passwords, and API keys
- Rug-pull drift where a previously approved server changes descriptor intent
- Semantic similarity to known MCP threat patterns

## Threats Out Of Scope For v0.1

- Executable binary malware analysis
- Full dependency/SBOM scanning
- Network sandboxing
- Live MCP protocol negotiation
- Runtime prompt-injection detection inside tool results
- Hosted team dashboards
- SSO or enterprise policy management

## Design Principles

- **Do not execute untrusted MCP servers by default.** v0.1 scans config and provided descriptor metadata only.
- **Keep private descriptors local.** The CLI works offline and does not upload tool names, descriptions, or env keys.
- **Prefer evidence over vibes.** Findings include rule IDs, severity, evidence, and recommendation.
- **Remember approved intent.** Lockfiles store descriptor fingerprints and vectors so drift is visible.
- **Treat semantic checks as advisory.** Similarity warnings should trigger review, not automatic condemnation.

## Risk IDs

| ID | Risk |
|---|---|
| `MCPCANARY-001` | Model-facing override instruction |
| `MCPCANARY-002` | Read-only claim conflicts with tool capability |
| `MCPCANARY-003` | Sensitive environment variable exposed to MCP server |
| `MCPCANARY-004` | Broad local or network capability |
| `MCPCANARY-005` | Hidden Unicode in descriptor |
| `MCPCANARY-006` | Descriptor resembles known MCP threat pattern |

## Recommended Use

Run MCPCanary before adding a server to Claude Desktop, Claude Code, Cursor, VS Code, Windsurf, Codex, or another MCP-capable client. Commit `mcpcanary.lock` for reviewed internal configs, then run `mcpcanary diff` in CI or before connecting a changed tool.
