# Requirements and traceability

Status: **Done** = implemented in `packages/core` with passing tests. **Planned** = needs a platform shell. **Open** = decision or provider documentation required.

## Accounts and providers

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-A1 | No provider is enabled or signed in on first run | Done | `accounts.ts` / `accounts.test.ts` |
| R-A2 | Per-provider status: disconnected, pending, authenticated, expired, unavailable | Done | `accounts.ts` |
| R-A3 | Sending requires enabled and authenticated; a provider error is shown as unavailable | Done | `accounts.ts` / `accounts.test.ts` |
| R-A4 | Disabling a provider keeps its data and does not sign out | Done | `accounts.ts` / `accounts.test.ts` |
| R-A5 | Sign-in uses only the provider's documented client | Open | Per-provider adapter; unsupported until its interface is documented |

## Workspace and editor

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-W1 | All file access is confined to the selected workspace, including via symlinks | Done | `paths.ts` / `paths.test.ts` |
| R-W2 | Dirty state tracked against the last known disk version | Done | `editor-buffer.ts` / `editor-buffer.test.ts` |
| R-W3 | Save detects external changes; user chooses overwrite or reload; nothing is lost silently | Done | `editor-buffer.ts` / `editor-buffer.test.ts` |
| R-W4 | Run profiles show command, working directory, arguments and environment before running | Planned | Shell UI |
| R-W5 | Search with case sensitivity, result count and next/previous | Planned | Shell UI |

## Terminal and background work

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-T1 | Terminal opens in the workspace directory; input, resize, interrupt, exit code, session list, stop-all | Planned | Platform adapters |
| R-T2 | Environment downloads show source, version, size and verify checksum or signature | Planned | Installer module |
| R-T3 | SSH only to user-authorised hosts with host-key verification | Planned | SSH module |

## Agent chat

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-C1 | Duplicate and out-of-order stream events are dropped after reconnect | Done | `chat-stream.ts` / `chat-stream.test.ts` |
| R-C2 | An interrupted or failed reply is never shown as complete | Done | `chat-stream.ts` / `chat-stream.test.ts` |
| R-C3 | Tool and command requests show target, effect and scope; allow once, allow scoped, allow always (confirmed), deny | Done | `permissions.ts` / `permissions.test.ts` |
| R-C4 | A grant does not apply to a different target, kind or workspace | Done | `permissions.test.ts` |
| R-C5 | Attachments are sent only when selected, with separate consent for sensitive files | Planned | Shell UI + `permissions.ts` (`attachment.send`) |
| R-C6 | HTML and SVG previews are treated as untrusted content | Planned | Renderer sandbox |

## Git

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-G1 | Git commands are built as argument arrays, never shell strings | Done | `git.ts` / `git.test.ts` |
| R-G2 | Paths must be inside the repository | Done | `git.ts` / `git.test.ts` |
| R-G3 | Discard, push and pull require confirmation | Done | `git.ts` / `git.test.ts` |
| R-G4 | Branch names cannot be parsed as options or revisions | Done | `git.ts` / `git.test.ts` |
| R-G5 | Status shows branch, staged, unstaged, untracked, ahead/behind; unknown state is not shown as empty | Planned | Status parser + UI |

## Files, backup and restore

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-F1 | Archives are validated before extraction: traversal, absolute paths, symlinks, duplicates (case and Unicode), entry count, size, disk space | Done | `archive.ts` / `archive.test.ts` |
| R-F2 | Extraction is transactional with rollback | Planned | Restore module |
| R-F3 | Secrets are excluded from backups by default | Planned | Backup module |
| R-F4 | Updates verify signature or checksum, version and architecture; user consent; rollback | Planned | Updater |

## Privacy and security

| ID | Requirement | Status | Implementation / test |
|---|---|---|---|
| R-S1 | Secrets are redacted from logs, exports and prompts | Done | `redact.ts` / `redact.test.ts` |
| R-S2 | Local helpers bind to loopback only and authenticate requests | Planned | Bridge module |
| R-S3 | Browser automation is off by default, shows a session indicator, and asks before submit, purchase, post or delete | Planned | Browser module |
| R-S4 | Telemetry, if any, is documented, opt-in and verifiable | Open | Decision required |

## Out of scope until separately specified

- YouTube playback (requires the official embed API and its terms)
- Cloud backup or sync
- Language servers, debugger, formatter
- Any use of undocumented provider endpoints, session cookies or tokens
