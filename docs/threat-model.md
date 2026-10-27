# Threat model

## Assets

- Source code and files in the user's workspaces
- Provider credentials and session state
- Git credentials and SSH keys
- Chat history, which may contain code, prompts and tool output
- The user's device: processes, storage, network

## Trust boundaries

1. User interface to agent. Agent output is untrusted input.
2. Core to operating system. Only adapters touch the file system, processes and network.
3. App to local helper (bridge, MCP worker, extension worker).
4. App to external services (AI providers, update server, Git remotes).
5. Imported data: archives, chat exports, extension packages, files opened from other apps.

## Threats and mitigations

| # | Threat | Mitigation | Verified by |
|---|---|---|---|
| T1 | Agent or file content causes a write or delete outside the workspace | `resolveInside` follows symlinks and rejects escapes | `paths.test.ts` |
| T2 | Agent runs a command without the user's knowledge | Default-deny `PermissionBroker`; the prompt shows target and effect | `permissions.test.ts` |
| T3 | A permission granted for one action is reused for another | Grants keyed by kind, target and workspace; expiry; revoke | `permissions.test.ts` |
| T4 | Shell injection through a commit message, path or branch name | Argument arrays; branch validation; `--` before paths | `git.test.ts` |
| T5 | Malicious archive writes outside the target or exhausts disk | Pre-extraction validation of names, links, duplicates, size, free space | `archive.test.ts` |
| T6 | Secrets leak through logs, crash reports, exports or prompts | `redact` on every outbound string; secrets excluded from backups | `redact.test.ts` |
| T7 | Partial reply treated as final | `ChatStream` state machine | `chat-stream.test.ts` |
| T8 | Overwriting a file changed elsewhere | Hash comparison on save; explicit overwrite or reload | `editor-buffer.test.ts` |
| T9 | A web page or local process calls the local helper | Loopback bind, per-request authentication, `Origin` validation, minimal scope | Planned (R-S2) |
| T10 | Tampered update or environment download | Signature or checksum verification, version pinning, rollback | Planned (R-F4, R-T2) |
| T11 | Untrusted extension or MCP server reads files or the network | Explicit permission scope per package; separate process; not described as a sandbox | Planned |
| T12 | Hostile HTML or SVG in chat executes script | Render in an isolated, script-disabled context | Planned (R-C6) |
| T13 | Account takeover through stored tokens | OS credential store; no tokens in config files, URLs or logs | Planned |

## Not mitigated by this design

- A compromised device or operating system
- Content the user explicitly approves
- Behaviour of third-party providers and MCP servers after the user connects them
