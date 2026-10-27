# Milestones

Each milestone ends with its acceptance tests passing on the platforms it targets.

| # | Scope | Exit criteria |
|---|---|---|
| M0 | Go bridge: guard, permission broker, Claude and agy headless backends, WebSocket server | `go vet` and `go test -race` pass. **Done** |
| M1 | Android app shell: project picker (SAF), file tree, editor with tabs and conflict-safe save, search | Acceptance B on a device |
| M2 | Terminal: native PTY sessions, foreground service, stop-all | Acceptance C on a device |
| M3 | Android environment: PRoot userland, verified downloads, install of `claude` and `agy` | Acceptance C and G |
| M4 | Chat UI: streaming, tool confirmation cards, questions, model and mode switching for Claude | Acceptance A and D |
| M5 | Git panel and bridge Git operations with confirmation | Acceptance F |
| M6 | Antigravity gRPC adapter, behind an opt-in setting | Traffic capture, then Acceptance D |
| M7 | Backup, restore, signed updates | Acceptance G |
| M8 | Desktop app | Acceptance A to H on each OS |
| M9 | MCP management, browser automation (opt-in) | Acceptance E and H |

Decisions needed: desktop framework for M8.
