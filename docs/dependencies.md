# Dependencies and licences

| Package | Version | Licence | Used for |
|---|---|---|---|
| typescript | ^5.9 | Apache-2.0 | Type checking (dev) |
| @types/node | ^24 | MIT | Node type definitions (dev) |

`packages/core` has no runtime dependencies. It uses Node.js built-ins (`node:fs`, `node:path`, `node:crypto`, `node:test`).

Each new dependency is added here with its licence and purpose before it is merged. Provider SDKs and CLIs are subject to the provider's current terms, which are checked before integration.
