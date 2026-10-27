# Architecture

## Layers

```
 Android shell          Desktop shell
 (touch UI, SAF,        (windows, keyboard, native
  foreground service)    pickers, OS credential store)
        \                      /
         \                    /
          platform adapters (per OS)
                  |
            packages/core
```

`packages/core` has no UI and no OS-specific code. It receives platform behaviour through small interfaces (`Disk`, `Prompter`, `ProviderAdapter`) and never calls a shell.

## Core modules

| Module | Responsibility |
|---|---|
| `paths` | Resolve a path inside a workspace; follow symlinks on the nearest existing ancestor; reject escapes |
| `permissions` | Default-deny broker. Grants are keyed by kind, target and workspace; they can expire and be revoked |
| `editor-buffer` | Separate in-memory content from the last known disk hash; detect conflicts on save |
| `chat-stream` | Order and de-duplicate stream events; mark interrupted or failed output as incomplete |
| `accounts` | Provider registry. Providers are opt-in; status comes from the provider's official client |
| `git` | Build `git` argv arrays. No shell strings; paths confined to the repository; confirmation required for destructive and remote operations |
| `archive` | Validate archive entries before extraction: traversal, symlinks, duplicates, size, disk space |
| `redact` | Remove secrets from any string before it is logged, exported or sent to a prompt |

## Platform adapters (planned)

| Concern | Android | Desktop |
|---|---|---|
| Workspace access | Storage Access Framework URI grants | OS folder picker, access limited to the chosen folder |
| Terminal | PTY session, foreground service while a task runs | OS shell, process-group cleanup on exit |
| Credentials | Android Keystore | Windows Credential Manager, macOS Keychain, libsecret |
| Provider sign-in | Provider's official browser flow | Provider's official CLI or browser flow |
| Updates | Signed APK handed to the system installer | Signed installer with rollback |

## Provider adapters

| Provider | Interface | State |
|---|---|---|
| Claude Code | Official CLI, `stream-json` input and output, one process per chat | Planned (M4 candidate) |
| Antigravity | None documented | Blocked. Enabled only after the provider documents a supported interface |

Adapters implement `ProviderAdapter` and feed events into `ChatStream`. Tool requests go to `PermissionBroker` before anything runs. Features such as live model switching are listed as supported only after an acceptance test passes against the installed CLI version.

## Process boundaries

- The UI process never executes agent-supplied commands directly. It asks the permission broker, then the process runner.
- Any local helper (bridge or daemon) binds to loopback only, authenticates each request, validates `Origin`, and exposes the minimum tool scope. It is not exposed on the LAN by default.
- Extension and MCP workers run as separate processes with their own permission scope. They are not presented as a security sandbox.
