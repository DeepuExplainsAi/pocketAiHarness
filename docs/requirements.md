# Requirements and traceability

**Done** means implemented with a passing test. **Planned** means it needs the app or a later bridge module. **Open** means a decision or provider documentation is required. Nothing is listed as supported until its acceptance test passes.

## Accounts and providers

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-A1 | No provider is enabled or signed in on first run | Planned | App |
| R-A2 | Per-provider status: disconnected, pending, authenticated, expired, unavailable | Planned | App, from each provider's own client |
| R-A3 | Sending is blocked unless the provider is authenticated | Planned | App |
| R-A4 | Disabling a provider keeps its data and does not sign out | Planned | App |
| R-A5 | Sign-in is done by the provider's own client; tokens are never read or stored | Done by design | `backend/` starts only the official binaries and never touches credentials |

## Workspace and editor

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-W1 | Agent file access is confined to the workspace, including through symlinks | Done | `guard/path.go`; `guard_test.go`; `backend/claude_test.go` |
| R-W2 | Dirty state tracked against the last known disk version | Planned | App |
| R-W3 | Save detects external changes; overwrite or reload is the user's choice | Planned | App |
| R-W4 | Run profiles show command, directory, arguments and environment first | Planned | App |
| R-W5 | Search with case option, result count, next and previous | Planned | App |

## Terminal and background work

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-T1 | Terminal in the workspace; input, resize, interrupt, exit code, sessions, stop-all | Planned | `android/` |
| R-T2 | Environment downloads show source, version and size, and verify a checksum | Planned | `android/` |
| R-T3 | SSH only to authorised hosts with host-key verification | Planned | `android/` |

## Agent chat

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-C1 | Duplicate or out-of-order events are dropped by the app using `seq` | Planned | App; the bridge numbers events |
| R-C2 | An interrupted or failed reply is never shown as complete | Planned | App |
| R-C3 | Tool requests show tool, description and input; allow or deny | Done (bridge) | `perm/`, `server/`; `server_test.go` round trip |
| R-C4 | An answer applies to one request only; a repeat is rejected | Done | `broker_test.go`, `server_test.go` |
| R-C5 | Disconnect, close, cancel and a missing approval channel deny | Done | `broker_test.go`, `claude_test.go` |
| R-C6 | A tool path outside the workspace is denied without asking | Done | `claude_test.go` |
| R-C7 | Claude permission mode and model change mid-chat | Done (bridge) | Verified against claude 2.1.311 |
| R-C8 | Antigravity tool approval | Open | Headless agy has none; needs the gRPC adapter |
| R-C9 | HTML and SVG in chat are untrusted | Planned | App renderer |

## Git

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-G1 | Commands are argument arrays, never shell strings | Planned | Bridge Git module (M5) |
| R-G2 | Paths must be inside the repository | Planned | M5, using `guard.ResolveInside` |
| R-G3 | Discard, push and pull require confirmation | Planned | M5 |
| R-G4 | Branch names cannot act as options or revisions | Planned | M5 |
| R-G5 | Status shows branch, staged, unstaged, untracked, ahead and behind; unknown is not shown as empty | Planned | M5 |

## Files, backup, restore

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-F1 | Archives are validated before extraction: traversal, absolute paths, symlinks, duplicates, size, disk space | Planned | M7 |
| R-F2 | Extraction is transactional with rollback | Planned | M7 |
| R-F3 | Secrets are excluded from backups by default | Planned | M7 |
| R-F4 | Updates verify signature or checksum, version and architecture, with rollback | Planned | M7 |

## Privacy and security

| ID | Requirement | Status | Where |
|---|---|---|---|
| R-S1 | Secrets are redacted from bridge errors | Done | `guard/redact.go`; `guard_test.go` |
| R-S2 | The bridge listens on loopback only and requires a token; `Origin` is rejected | Done | `server_test.go`; checked against the built binary |
| R-S3 | Browser automation is off by default, with a session indicator and confirmation before consequential actions | Planned | M9 |
| R-S4 | Telemetry, if any, is documented and opt-in | Open | Decision |

## Out of scope until separately specified

- YouTube playback (official embed API and its terms)
- Cloud backup or sync
- Language servers, debugger, formatter
- Reading provider tokens, session cookies or credential stores
