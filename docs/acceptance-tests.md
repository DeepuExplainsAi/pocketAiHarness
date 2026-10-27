# Acceptance tests

Run every item separately on each supported platform. A feature is listed as supported only after its items pass.

## A. Install, setup, accounts
- [ ] A clean install has no provider enabled or signed in
- [ ] Enabling or disabling a provider matches the real service state; history is kept
- [ ] Sign-in: success, cancel, deny, expiry, offline, logout. No credential in logs, URLs or analytics
- [ ] Missing CLI, version mismatch, install failure and unsupported account are reported clearly
- [ ] Permission denial (Android runtime, desktop OS) degrades gracefully

## B. Workspace and editor
- [ ] Open a project; open a deep folder; create, rename, delete; read-only and permission errors
- [ ] Dirty state, undo/redo, tab switch, save, close with save/discard/cancel
- [ ] External disk change: overwrite and reload both behave correctly; no silent loss
- [ ] Search with case options and cyclic navigation
- [ ] Binary, large, UTF-8, line endings, SVG, image, invalid or revoked URI
- [ ] Run with an unsaved buffer prompts; command, directory, arguments and environment are shown

## C. Terminal and background work
- [ ] Session opens in the workspace; input, resize, interrupt, exit code
- [ ] Stop-all ends every process; no orphans after the app closes
- [ ] Background task survives the OS lifecycle on Android; user sees a running indicator
- [ ] Environment download verifies checksum or signature; failure rolls back
- [ ] SSH rejects an unknown or changed host key

## D. Agent chat
- [ ] Status is shown before sending; unauthenticated provider blocks sending
- [ ] Stream interrupted: partial reply marked incomplete; no duplicate message on retry
- [ ] Permission prompt shows target, effect and scope; once, scoped, always, deny all work
- [ ] Denied or timed-out request leaves state consistent after refresh
- [ ] HTML and SVG content cannot run script or reach the network
- [ ] Export warns that prompts and tool history are included

## E. MCP and extensions
- [ ] Add, edit, remove, enable, disable, refresh; transport, tools and scope are shown before connecting
- [ ] Secrets are never stored in plain config or logs
- [ ] Removing a connector stops its process and clears its credentials as described

## F. Git
- [ ] Non-repository folder, detached HEAD, no remote, no upstream
- [ ] Status separates staged, unstaged, untracked, ahead and behind; unknown is not shown as empty
- [ ] Stage-all requires a visible selection; discard, checkout, push, pull and stash-pop show the effect first
- [ ] Hostile file names and branch names do not alter the command
- [ ] Credentials come from the Git credential helper or SSH; none shown in the UI

## G. Backup, restore, update
- [ ] Included and excluded content listed; secrets excluded by default
- [ ] Hostile archive (traversal, symlink, duplicate, oversized) is rejected before any file is written
- [ ] Low disk, cancel, partial failure: the target is restored to its prior state
- [ ] After restore, file hashes match and accounts ask to sign in again
- [ ] Updater: invalid metadata, missing asset, bad signature, downgrade, user rejects; rollback works

## H. Privacy, security, quality
- [ ] Telemetry matches the published policy; opt-out works
- [ ] The bridge listens on loopback only; a request without the token, or with an `Origin` header, gets 403
- [ ] Permissions are requested only when needed, with a stated reason
- [ ] Screen reader, keyboard-only navigation (desktop), contrast, scaling, touch target size
- [ ] Crash logs contain no secrets
