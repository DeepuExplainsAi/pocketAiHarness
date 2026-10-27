# Pocket AI Harness

A development workspace for Android phones and desktop PCs: edit files, run a terminal, use Git, and work with an AI coding agent, with every consequential action gated by explicit user consent.

Status: pre-alpha. The shared core is implemented and tested. The Android and desktop shells are planned (see [milestones](docs/milestones.md)).

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
