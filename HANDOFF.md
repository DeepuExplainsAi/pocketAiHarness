# Pocket AI Harness: hand-off

For the next engineer or AI. Everything here was checked on 2026-10-27 against commit `23aec40`. Where something was not verified, it says so.

Product: an AI coding harness for phone (Android) and PC. Name is always **Pocket AI Harness**. Repo: https://github.com/DeepuExplainsAi/pocketAiHarness (public, branch `main`).

Owner writes Hinglish. Reply in the owner's language, keep text plain and professional (no marketing wording, no emoji).

---

## 1. Current state in one table

| Area | State | Evidence |
|---|---|---|
| Go bridge (`bridge/`): workspace path guard, secret redaction, permission broker, loopback WebSocket server | Done | `go vet ./... && go test -race ./...` pass |
| Claude Code backend | Done, partly live-verified | Control messages checked against the real `claude` 2.1.311 (`set_permission_mode`, `set_model`, `interrupt`, `can_use_tool` fields). The full tool-approval round trip was tested with a fake CLI only, never with a signed-in Claude |
| Antigravity headless backend (`agy -p` stream-json) | Written, never run end to end | Flags read from `agy --help` (1.4.8). Needs a Google sign-in to run |
| Antigravity gRPC backend | **Not started** | Protocol undocumented, see section 4 |
| Android app (`android/`): chat screen, permission cards, Stop, bridge client | Builds, 14 unit tests pass | Never installed or run on a device. Bridge address and token are typed in by hand |
| Android: PRoot Linux userland, launching the bridge, terminal, folder access, editor, Git, backup, updater | **Not started** | |
| Desktop app | **Not started**, framework undecided | |
| MCP management, browser automation | **Not started** | |

---

## 2. What exists (files)

```
README.md                  overview, feature table, architecture
HANDOFF.md                 this file
LICENSE                    MIT
docs/architecture.md       components, backends, process boundaries
docs/protocol.md           app <-> bridge WebSocket messages
docs/requirements.md       every requirement with Done / Partly / Planned / Open
docs/threat-model.md       threats and mitigations
docs/milestones.md         M0 to M9
docs/acceptance-tests.md   acceptance tests A to H (from the owner's spec)
docs/dependencies.md       dependencies and licences
bridge/                    Go module
  cmd/pah-bridge/main.go     entry; prints {"listen","token"} on start
  internal/guard/            ResolveInside (symlink-safe), Redact
  internal/perm/             permission broker (default deny)
  internal/backend/          Session interface, claude.go, agy.go
  internal/server/           loopback WebSocket, bearer token, Origin rejected
android/                   Kotlin + Jetpack Compose app
  app/src/main/java/ai/pocketharness/app/
    Protocol.kt              parse/build bridge messages, never throws
    ChatState.kt             reducer: seq dedupe, gap = lost data, interrupted/failed replies
    BridgeClient.kt          OkHttp WebSocket with bearer token
    ChatViewModel.kt         one chat per connection
    ChatScreen.kt            chat list, permission cards, Stop
    MainActivity.kt          manual address/token/workspace entry
  app/src/test/.../ChatStateTest.kt   14 tests
```

---

## 3. Behaviour that is implemented and tested

- Bridge binds only to a loopback address; any other address is refused at start.
- WebSocket needs `Authorization: Bearer <token>`. No token, wrong token, or any `Origin` header gives 403.
- Claude runs as `claude -p --input-format stream-json --output-format stream-json --verbose --include-partial-messages --permission-prompt-tool stdio --setting-sources "" --permission-mode manual`. User, project and local settings are off so allow rules and hooks cannot skip the prompt. Modes that auto-approve are refused at start. `SetMode` accepts only `manual` and `plan`.
- Each Claude `can_use_tool` request becomes a `permission` message to the app. Only a `decide` message allows it. Deny on: app disconnect, chat close, provider exit, no approval channel, repeat answer, file path outside the workspace, or an unexpected path value shape (the last two are denied without asking).
- `ResolveInside` follows symlinks on the nearest existing ancestor and does not fold `..` after a link.
- App marks a reply interrupted or failed rather than complete when the stream closes, errors, has a `seq` gap, or the provider output could not be read. Permission cards are cleared on close or disconnect. A full inbound buffer closes the socket instead of dropping messages.
- Malformed JSON from the bridge never throws in the app.

---

## 4. Not done, and why

### 4.1 Antigravity gRPC (owner wants this on the phone)
- Owner's spec: connect directly to the Antigravity language server over gRPC from the app, with live streaming, sub-agents, questions and tool confirmations.
- Found: `agy` 1.4.8 (installer: `curl -fsSL https://antigravity.google/cli/install.sh | bash`, needs bash) has a `--hub-port <port>` flag text and an `antigravity-hub` auth mode, and the installer supports `linux_arm64` and detects Android/proot. The hub wire protocol is not documented and was not observed, because `agy` stops at "Authentication required" without a Google login.
- Public community schemas: `jkfujinami/antigravity-grpc-schemas` (MIT, 39 `.proto`, reverse-engineered from the IDE bundle, Antigravity 2.18.1). Services include `LanguageServerService`, `ChatClientServerService`. Unofficial and may break on any update. Their repo asks for credit; add it to `docs/dependencies.md` if used.
- Do not copy the community SDK's standalone mode: it reads OAuth tokens from the IDE's `state.vscdb`. Pocket AI Harness must never read, copy or store provider tokens.
- Google forum reply (2026-09-16, responder not confirmed as Google staff): spawning the unmodified `agy` binary as a single-user child process under the user's own login is compliant; the ToS targets credential harvesting, unofficial API reverse-engineering and multi-tenant proxying. Google also ships an official Antigravity ACP server.
- Decision taken with the owner: build gRPC as an **opt-in** adapter, keep the headless/ACP path as the default and fallback.
- Next steps: (1) signed-in `agy` session, start it with `--hub-port`, capture what it serves; (2) generate Go stubs from the schemas for only the services needed; (3) implement `backend.Session` for it; (4) map streaming, questions and tool confirmations to the existing `perm` broker; (5) add the acceptance test.
- Headless `agy` cannot ask for tool approval (unapproved tools are soft-denied; `control_request` is rejected). The bridge does not check agy's own settings, so tool use through agy is **unconfirmed**.

### 4.2 Android
- Run the app on a real device. Never done.
- Bridge launch from the app (today the user starts `pah-bridge` and types the address and token).
- PRoot Linux userland (own environment, no Termux): rootfs download with source, size, version and pinned checksum shown before use; install `claude` and `agy` inside it.
- Native PTY terminal, foreground service, stop-all.
- Folder access (SAF), file tree, editor with tabs, conflict-safe save, search.
- Git panel, backup and restore, signed updates.
- Release signing. The existing APK is an unsigned debug build.

### 4.3 Bridge
- OS sandbox for approved commands. Today the bridge cannot confine what an approved shell command does or what URL a tool fetches; only the approval card protects those. Not verified that every Claude tool and MCP server goes through the permission callback.
- Provider events and permission payloads are forwarded unredacted.
- Reconnect starts a new provider session with a new chat id; it does not resume.
- Mode and model change errors from Claude are not reported back (the `control_response` is discarded).
- No permission timeout.
- Git module (argument arrays, branch validation, path check), archive validation, update verification.

### 4.4 Other
- Desktop app and framework choice.
- MCP management and browser automation (opt-in, with session indicator and confirmation).
- YouTube playback (official embed API only), cloud sync: out of scope until separately specified.
- Telemetry policy: undecided; if any, opt-in and documented.

Milestone order is in `docs/milestones.md`.

---

## 5. What the owner must provide or decide

1. **A signed-in Antigravity session** to capture the `--hub-port` traffic (the owner signs in; the assistant must never take the token). Without it, gRPC cannot be built.
2. **Desktop framework**: Tauri or Electron. The bridge is Go, so either works.
3. **A real Android phone** (or emulator access) and its CPU type, to test the APK and PRoot. Android 8 (API 26) or newer.
4. **Anthropic sign-in** for Claude Code, done by the owner in the provider's own client, to test the real approval round trip.
5. **Telemetry**: none, or opt-in only.
6. **GitHub**: the `gh` token has scopes `gist, read:org, repo`. Building the APK in GitHub Actions needs the `workflow` scope, so the owner must run `gh auth refresh -s workflow`, or add the workflow file in the web UI.
7. **Android release signing key** for a release APK (keep it out of the repo).
8. **Tell the owner if the provider rules matter**: whether private Antigravity interfaces are acceptable for their account. The original spec forbade private protocols; the latest instructions asked for gRPC. This is the owner's call and risk.

---

## 6. How to build and test

Bridge (Go 1.27):
```sh
cd bridge
go vet ./... && go test -race ./...
go build ./cmd/pah-bridge
./pah-bridge            # prints {"listen":"127.0.0.1:PORT","token":"..."}
```
Cross-build, no cgo: `GOOS=linux GOARCH=arm64`, `GOOS=android GOARCH=arm64`, `windows/amd64`, `darwin/arm64` all compiled (about 7 MB).

Android (JDK 17, Android SDK 35 with build-tools 35.0.0, Gradle 8.14.3, about 1.5 GB RAM):
```sh
cd android
echo "sdk.dir=/path/to/sdk" > local.properties     # not committed
gradle --no-daemon :app:testDebugUnitTest :app:assembleDebug < /dev/null
```
Debug APK: `android/app/build/outputs/apk/debug/app-debug.apk`.

Push: `gh auth setup-git`, then `GIT_TERMINAL_PROMPT=0 git push origin main`.

### Sandbox notes (the machine this was built on)
- 482 MB RAM and a 2 GB `/root` disk. A swap file on `/` (`/opt/swap`) and `GRADLE_USER_HOME=/opt/gh` were needed. Put big files outside `/root` or `/root` fills and tools fail with misleading errors.
- Set `JAVA_TOOL_OPTIONS=-Duser.home=/opt/gh`. Run Gradle in the background with stdin from `/dev/null`; in the foreground it stops on a TTY read.
- Gradle failing with "daemon disappeared" means out of memory. A "merged_manifest ... doesn't exist" error came from two Gradle processes running at once; kill all and retry.
- `zip` is missing; use Python `shutil.make_archive`.
- Tool output truncates long lines; wrap with `textwrap` before reading specs.

---

## 7. Rules for whoever continues

- Do not mark anything Done until it has a passing test; update `docs/requirements.md` in the same commit.
- Consent first: every consequential agent action needs the user's explicit answer in the app. Agent output is never consent.
- Never read, copy, log or store provider tokens, cookies or credential stores. Never take them from `state.vscdb` or similar.
- Redact secrets before logging or sending errors.
- New dependency: add it with its licence to `docs/dependencies.md` first.
- Verify protocol claims against the real binary (`--help`, running it), not memory. Several earlier guesses were wrong until checked.

## 8. History

- `fe65490` initial commit. `703b617` first TypeScript core (later removed). `24bbdeb` README features. `05d982a` Go bridge replaced the TypeScript core. `06ff000` Android chat screen. `23aec40` hardening after an independent review.
- The original owner spec (Hindi, clean-room behaviour only) forbade private provider protocols and token extraction; the owner's later instructions asked for direct Antigravity gRPC on the phone. Both are recorded; the safe parts of the first still apply.
