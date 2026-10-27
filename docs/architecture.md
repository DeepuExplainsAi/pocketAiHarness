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
| agy headless | `agy -p --input-format stream-json --output-format stream-json`, resume with `--conversation` | Implemented. No approval channel; agy soft-denies unapproved tools |
| agy gRPC | Language-server connection | Planned, after traffic capture from a signed-in session |

## Android environment

The app ships its own Linux userland so Termux is not required. It runs inside the app's private storage through PRoot. Android's `root` inside PRoot is not superuser. Downloads (rootfs, packages, the `agy` installer, the `claude` CLI) are shown with source, size and version, and checked against a pinned checksum before use.

## Process boundaries

- The UI never runs agent commands itself. The bridge asks the user through `perm`.
- The bridge listens on loopback only, requires a per-run bearer token and rejects any request with an `Origin` header.
- Secrets pass through `Redact` before reaching logs or error messages.
- Provider tokens are never read, copied or stored by Pocket AI Harness. Sign-in is done by the provider's own client.
