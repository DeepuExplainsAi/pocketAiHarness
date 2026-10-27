# Pocket AI Harness

A mobile and desktop IDE with an integrated AI coding agent, a Git client and a full terminal. It runs on Android phones and on PCs from one shared core.

Status: pre-alpha. The Go bridge is implemented and tested for Claude Code. The Android app has a chat screen with permission cards and builds, but has not been run on a device and does not start the bridge itself. The Antigravity gRPC adapter and the desktop app are not built yet (see [milestones](docs/milestones.md)).

## Features

| Area | Description | Status |
|---|---|---|
| Editor | Tabs, search, undo/redo, dirty tracking, conflict-safe save | Planned |
| Agent chat | Streaming replies, tool confirmations, per-chat provider | Bridge done for Claude Code; app planned |
| Terminal | PTY sessions in the workspace, background tasks for long builds and servers | Planned |
| Git | Status, diff, stage, commit, branches, push and pull with confirmation | Planned |
| Standalone environment | Own Linux userland and package manager inside the app sandbox, no Termux required | Planned (Android) |
| Backup and restore | Validated, transactional, secrets excluded by default | Planned |

## Architecture

```
  Android app (Kotlin, Compose)          Desktop app
  touch UI, native PTY, own Linux        windows, keyboard, native shell
              \                              /
               +---- WebSocket, loopback ---+
                            |
                   pah-bridge (Go)
      token auth, workspace guard, permission broker
                            |
          +-----------------+-----------------+
          |                 |                 |
   Claude Code CLI     agy stream-json    agy gRPC
   stream-json         (headless)         (planned)
   one process/chat    no tool approval
```

The bridge owns permissions, path containment and secret redaction. The apps supply the UI and the operating-system pieces. See [docs/protocol.md](docs/protocol.md).

## Agent integration

| Backend | Interface | State |
|---|---|---|
| Claude Code | Official `claude` CLI, `stream-json` in and out, `--permission-prompt-tool stdio` | Implemented, verified against claude 2.1.311 |
| Antigravity, headless | Official `agy -p --input-format stream-json --output-format stream-json` | Implemented. Cannot ask for tool approval; agy denies unapproved tools itself |
| Antigravity, gRPC | Direct connection to the language server | Not implemented. See below |

**Claude Code.** Tool requests arrive as `control_request` messages and are answered with `control_response`. Permission mode and model change mid-chat through `set_permission_mode` and `set_model`; both were confirmed against the installed CLI.

**Antigravity gRPC.** Public schemas are reverse-engineered, unofficial and tied to Antigravity 2.18.1. The installed `agy` 1.4.8 contains a `--hub-port` flag and an `antigravity-hub` auth mode, but the hub protocol is not documented, and it could not be exercised without a signed-in account. The adapter will be written against captured traffic from a real session. Google's forum guidance treats spawning the official binary under the user's own login as compliant and warns against token extraction and unofficial API reverse-engineering, so the gRPC path is opt-in and the headless path stays available.

**Consent.** A tool request from any agent is shown to the user with its target. Only the app can approve it.

## Principles

- **Consent first.** File writes, deletes, command execution, Git pushes, browser submissions and attachment uploads require a decision from the user interface. Agent output is never consent.
- **Official interfaces only.** AI providers are used through their documented CLIs, SDKs and APIs, with the normal user-authorised sign-in. No token extraction, no private protocols.
- **Scoped by default.** Everything is confined to the workspace the user selected. Paths are resolved through symlinks before use.
- **No silent data loss.** Unsaved buffers, disk conflicts, interrupted streams and failed restores are surfaced, not hidden.
- **Honest status.** A feature is listed as supported only when its acceptance test passes.

## Repository layout

| Path | Contents |
|---|---|
| `android/` | Kotlin app: protocol client, chat state, Compose chat screen with permission cards |
| `bridge/` | Go bridge: `guard` (workspace paths, redaction), `perm` (permission broker), `backend` (Claude, agy), `server` (WebSocket) |
| `docs/protocol.md` | App to bridge messages |
| `docs/requirements.md` | Requirements, each mapped to an implementation and a test |
| `docs/architecture.md` | Shared core, platform adapters, process boundaries |
| `docs/threat-model.md` | Assets, trust boundaries, threats and mitigations |
| `docs/milestones.md` | Phased delivery plan |
| `docs/acceptance-tests.md` | Per-platform acceptance checklist |
| `docs/dependencies.md` | Dependency and licence inventory |

## Development

Requires Go 1.27 or newer.

```sh
cd bridge
go vet ./...
go test -race ./...
go build ./cmd/pah-bridge
```

Android (JDK 17, Android SDK 35, Gradle 8.14, about 1.5 GB of RAM):

```sh
cd android
gradle :app:testDebugUnitTest :app:assembleDebug
```

## Licence

MIT. See [LICENSE](LICENSE).
