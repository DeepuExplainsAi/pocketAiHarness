# Milestones

Each milestone ends with its acceptance tests passing on every platform it targets.

| # | Scope | Platforms | Exit criteria |
|---|---|---|---|
| M0 | Shared core: paths, permissions, editor buffer, chat stream, accounts, Git args, archive validation, redaction | Both | Typecheck and tests pass. **Done** |
| M1 | Desktop shell: open folder, tabs, editor, save with conflict handling, search | Desktop | Acceptance section B |
| M2 | Terminal and run profiles; process lifecycle; stop-all | Desktop, then Android | Acceptance section C |
| M3 | Git panel: status parser, stage, commit, diff, confirmed destructive actions | Both | Acceptance section F |
| M4 | Agent chat with one provider through its documented interface; permission prompts; streaming states | Both | Acceptance sections A and D |
| M5 | Android shell: SAF workspace access, touch editor, foreground service | Android | Acceptance sections A to D on a device |
| M6 | Backup, restore, signed updates | Both | Acceptance section G |
| M7 | MCP connectors and extensions with scoped permissions | Both | Acceptance section E |
| M8 | Browser automation (opt-in) | Both | Browser items in acceptance sections D and H |

Open decisions before M1: desktop framework (Electron or Tauri), Android stack (Kotlin native or shared web UI), and the first provider to integrate in M4.
