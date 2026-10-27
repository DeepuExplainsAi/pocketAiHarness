# Dependencies and licences

| Component | Version | Licence | Used for |
|---|---|---|---|
| Go toolchain | 1.27 | BSD-3-Clause | Build |
| github.com/coder/websocket | 1.8.15 | ISC | Bridge WebSocket server |

The bridge has no other third-party dependencies.

Android app (Gradle):

| Component | Version | Licence |
|---|---|---|
| Android Gradle Plugin | 8.9.3 | Apache-2.0 |
| Kotlin | 2.1.21 | Apache-2.0 |
| Compose BOM / Material 3 | 2025.06.00 | Apache-2.0 |
| AndroidX Activity, Lifecycle | 1.10.1, 2.9.1 | Apache-2.0 |
| kotlinx-serialization-json | 1.8.1 | Apache-2.0 |
| kotlinx-coroutines | 1.10.2 | Apache-2.0 |
| OkHttp | 4.12.0 | Apache-2.0 |

Runtime tools that the app installs for the user, under their own terms:

| Tool | Source | Terms |
|---|---|---|
| Claude Code CLI | Anthropic | Anthropic terms |
| Antigravity CLI (`agy`) | Google | Antigravity terms |

If the gRPC adapter reuses the community schemas (`jkfujinami/antigravity-grpc-schemas`, MIT), the repository asks for credit; that notice will be added here. Those schemas are reverse-engineered and unofficial.

Each new dependency is listed here with its licence before it is merged.
