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
| T1 | Agent or file content causes a write or delete outside the workspace | `guard.ResolveInside` follows symlinks and rejects escapes | `guard/guard_test.go`, `backend/claude_test.go` |
| T2 | Agent runs a command without the user's knowledge | Default-deny broker; an explicit answer is required | `perm/broker_test.go`, `server/server_test.go` |
| T3 | A permission granted for one action is reused for another | One answer per request id | `perm/broker_test.go`, `server/server_test.go` |
| T4 | Shell injection through a commit message, path or branch name | Argument arrays; branch validation; `--` before paths (M5) | Planned (M5) |
| T5 | Malicious archive writes outside the target or exhausts disk | Pre-extraction validation of names, links, duplicates, size, free space | Planned (M7) |
| T6 | Secrets leak through logs, crash reports, exports or prompts | `guard.Redact` on bridge errors; secrets excluded from backups (planned, M7); provider events and permission payloads are forwarded unredacted | `guard/guard_test.go` |
| T7 | Partial reply treated as final | App state machine using `seq` | Planned (app) |
| T8 | Overwriting a file changed elsewhere | Hash comparison on save (app) | Planned (app) |
| T9 | A web page or local process calls the local helper | Loopback bind, bearer token, `Origin` rejected | `server/server_test.go` |
| T10 | Tampered update or environment download | Signature or checksum verification, version pinning, rollback | Planned (R-F4, R-T2) |
| T11 | Untrusted extension or MCP server reads files or the network | Explicit permission scope per package; separate process; not described as a sandbox | Planned |
| T12 | Hostile HTML or SVG in chat executes script | Render in an isolated, script-disabled context | Planned (R-C6) |
| T13 | Account takeover through stored tokens | OS credential store; no tokens in config files, URLs or logs | Planned |

## Not mitigated by this design

- A compromised device or operating system
- Content the user explicitly approves
- Behaviour of third-party providers and MCP servers after the user connects them
