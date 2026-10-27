# Pocket AI Harness

A mobile and desktop IDE with an integrated AI coding agent, a Git client and a full terminal. It runs on Android phones and on PCs from one shared core.

Status: pre-alpha. The shared core is implemented and tested. The Android and desktop shells are planned (see [milestones](docs/milestones.md)).

## Features

| Area | Description | Status |
|---|---|---|
| Editor | Tabs, search, undo/redo, dirty tracking, conflict-safe save | Core done, UI planned |
| Agent chat | Streaming replies, tool confirmations, interactive questions, per-chat provider | Core done, UI planned |
| Terminal | PTY sessions in the workspace, background tasks for long builds and servers | Planned |
| Git | Status, diff, stage, commit, branches, push and pull with confirmation | Argument builder done, UI planned |
| Standalone environment | Own Linux userland and package manager inside the app sandbox, no Termux required | Planned (Android) |
| Backup and restore | Validated, transactional, secrets excluded by default | Validation done, rest planned |

## Architecture

```
  Android app                       Desktop app
  (touch UI, native PTY)            (windows, keyboard, native shell)
          \                              /
           +---------- shared core ------+
                         |
              local agent bridge
        (loopback only, authenticated requests)
                         |
          +--------------+--------------+
          |                             |
  Claude Code CLI adapter       Antigravity adapter
  (stream-json, one process     (blocked until a documented
   per chat)                     interface is available)
```

The core owns permissions, path containment, editor state, stream handling and Git command construction. Platform shells supply the UI and the operating-system pieces.

## Agent integration

Agents are reached through each provider's documented interface, not by reading terminal output.

- **Claude Code.** The bridge runs the official `claude` CLI in `stream-json` mode and relays events to the app. This gives incremental output, tool-use events and permission prompts without screen scraping. Model, effort and permission mode are meant to change during a chat without restarting it, subject to what the CLI supports; this is verified by acceptance tests before it is advertised.
- **Antigravity.** The adapter is not started until the provider publishes a documented, supported interface. Pocket AI Harness does not reverse-engineer or depend on private protocol schemas.
- **Consent.** A tool or permission request from any agent is shown to the user with its target and effect. Only the user interface can approve it.

## Principles

- **Consent first.** File writes, deletes, command execution, Git pushes, browser submissions and attachment uploads require a decision from the user interface. Agent output is never consent.
- **Official interfaces only.** AI providers are used through their documented CLIs, SDKs and APIs, with the normal user-authorised sign-in. No token extraction, no private protocols.
- **Scoped by default.** Everything is confined to the workspace the user selected. Paths are resolved through symlinks before use.
- **No silent data loss.** Unsaved buffers, disk conflicts, interrupted streams and failed restores are surfaced, not hidden.
- **Honest status.** A feature is listed as supported only when its acceptance test passes.

## Repository layout

| Path | Contents |
|---|---|
| `packages/core` | Platform-neutral logic: path containment, permission broker, editor buffer, chat stream, provider registry, Git argument builder, archive validation, secret redaction |
| `docs/requirements.md` | Requirements, each mapped to an implementation and a test |
| `docs/architecture.md` | Shared core, platform adapters, process boundaries |
| `docs/threat-model.md` | Assets, trust boundaries, threats and mitigations |
| `docs/milestones.md` | Phased delivery plan |
| `docs/acceptance-tests.md` | Per-platform acceptance checklist |
| `docs/dependencies.md` | Dependency and licence inventory |

## Development

Requires Node.js 24 or newer.

```sh
npm install
npm run typecheck
npm test
```

## Licence

MIT. See [LICENSE](LICENSE).
