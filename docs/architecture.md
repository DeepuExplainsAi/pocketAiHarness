# Architecture

## Components

```
 Android app (Kotlin, Jetpack Compose)        Desktop app
 UI, native PTY terminal, foreground          UI, OS shell, native pickers
 service, own Linux userland (PRoot)
              \                                  /
               +------ WebSocket, loopback -----+
                              |
                    pah-bridge (Go)
                              |
        +---------------------+---------------------+
        |                     |                     |
  Claude Code CLI       agy headless          agy gRPC (planned)
```

| Component | Language | Responsibility |
|---|---|---|
| `bridge/` | Go | Sessions, permission broker, workspace guard, redaction, WebSocket server |
| `android/` | Kotlin | UI, terminal, SAF workspace access, foreground service, environment installer |
| `desktop/` | To be decided | UI and native integration |

Go compiles to one static binary per target (`linux/arm64` for the phone userland, `android/arm64`, `windows/amd64`, `darwin/arm64`), so the same bridge runs on phone and PC.

## Bridge packages

| Package | Responsibility |
|---|---|
| `guard` | `ResolveInside` confines paths to the workspace, following symlinks on the nearest existing ancestor. `Redact` removes secrets from text |
| `perm` | Request broker. Allow only on an explicit answer. Disconnect, cancel, unknown id and repeated answer all deny |
| `backend` | `Session` interface with Claude and agy implementations |
| `server` | Loopback-only WebSocket with bearer token and `Origin` rejection |

## Provider backends

| Backend | Interface | State |
|---|---|---|
| Claude Code | `claude -p --input-format stream-json --output-format stream-json --verbose --include-partial-messages --permission-prompt-tool stdio`. Approvals via `control_request` and `control_response`. Mode and model switch via `set_permission_mode` and `set_model` | Implemented |
| agy headless | `agy -p --input-format stream-json --output-format stream-json`, resume with `--conversation` | Implemented, not run end to end (needs sign-in). The bridge has no approval channel for it and does not verify agy's own permission settings, so treat it as unconfirmed for tool use |
| agy gRPC | Language-server connection | Planned, after traffic capture from a signed-in session |

## Android environment (planned, milestone M3)

The app will ship its own Linux userland so Termux is not required. It runs inside the app's private storage through PRoot. Android's `root` inside PRoot is not superuser. Downloads (rootfs, packages, the `agy` installer, the `claude` CLI) are shown with source, size and version, and checked against a pinned checksum before use.

## Process boundaries

- The UI never runs agent commands itself. The bridge asks the user through `perm`.
- Known limits: the bridge cannot confine what an approved shell command does, cannot see URLs a tool fetches, and trusts the provider to send every tool use through its permission callback. An OS sandbox is planned. Until then the approval card is the control.
- The bridge listens on loopback only, requires a per-run bearer token and rejects any request with an `Origin` header.
- Secrets pass through `Redact` before reaching logs or error messages.
- Provider tokens are never read, copied or stored by Pocket AI Harness. Sign-in is done by the provider's own client.
