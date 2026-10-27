# App to bridge protocol

The app talks to `pah-bridge` over one WebSocket at `ws://127.0.0.1:<port>/ws`.

## Connection
- The bridge prints one JSON line on start: `{"listen":"127.0.0.1:41815","token":"..."}`.
- The app connects with `Authorization: Bearer <token>`. A request with no token, a wrong token, or an `Origin` header is refused with 403.
- The bridge refuses to listen on any non-loopback address.

## App to bridge
All messages are JSON objects with an `op` field.

| op | Fields | Effect |
|---|---|---|
| `start` | `chat`, `provider` (`claude` or `agy`), `cwd`, optional `model`, `mode`, `resume` | Start one session for the chat |
| `send` | `chat`, `text` | Send a user message |
| `interrupt` | `chat` | Stop the current turn |
| `mode` | `chat`, `mode` | Change permission mode mid-chat (Claude) |
| `model` | `chat`, `model` | Change model mid-chat (Claude) |
| `decide` | `request`, `allow`, optional `reason` | Answer a permission request |
| `close` | `chat` | End the session |

## Bridge to app
| type | Fields |
|---|---|
| `started` | `chat` |
| `event` | `chat`, `seq`, `raw` (provider message, unchanged) |
| `permission` | `chat`, `request`, `tool`, `description`, `input` |
| `error` | `chat`, `message` (secrets removed) |
| `closed` | `chat` |

`seq` increases by one per chat. The app drops any event with a `seq` it has already seen.

## Permission rules
- Every tool request is shown to the user. Only a `decide` message can allow it.
- A request is denied when the app disconnects, when the chat closes, when no approval channel exists, and when a path input resolves outside `cwd`. The last case is denied without asking.
- A request id can be answered once. A second answer returns `error`.
