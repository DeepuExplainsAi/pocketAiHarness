# Dependencies and licences

| Component | Version | Licence | Used for |
|---|---|---|---|
| Go toolchain | 1.27 | BSD-3-Clause | Build |
| github.com/coder/websocket | 1.8.15 | ISC | Bridge WebSocket server |

The bridge has no other third-party dependencies.

Runtime tools that the app installs for the user, under their own terms:

| Tool | Source | Terms |
|---|---|---|
| Claude Code CLI | Anthropic | Anthropic terms |
| Antigravity CLI (`agy`) | Google | Antigravity terms |

If the gRPC adapter reuses the community schemas (`jkfujinami/antigravity-grpc-schemas`, MIT), the repository asks for credit; that notice will be added here. Those schemas are reverse-engineered and unofficial.

Each new dependency is listed here with its licence before it is merged.
