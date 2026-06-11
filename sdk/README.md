# Rundown-Workers SDK Hub

This directory contains the language-specific SDKs for Rundown-Workers. Each SDK acts as a wrapper around the Rundown-Workers Engine API.

## Available SDKs

| Language | Status | Directory |
| :--- | :--- | :--- |
| **Python** | Beta (v0.1.1) | [`python/`](./python) |
| **Go** | Beta (v0.1.0) | [`go/`](./go) |
| **Node.js** | Planned | [`nodejs/`](./nodejs) |

## SDK Design Philosophy

All Rundown-Workers SDKs should provide a highly developer-friendly interface, typically using decorators (for registration) and a centralized runner.

### Core Protocol Requirements

To maintain consistency across languages, every SDK should implement:

1.  **Registration**: A way to map a function to a queue name (e.g., `@rw.queue(name="my-queue")`).
2.  **Concurrency-safe Polling**: The worker must continuously poll the `/poll` endpoint.
3.  **Timeout Enforcement**: Every SDK should locally track the execution time of a task and report failure to `/fail` if the task hangs beyond its defined `timeout`.
4.  **Graceful Shutdown**: Workers should catch termination signals (SIGINT/SIGTERM) and stop polling before exiting.

---

## Contributing a New SDK

If you're building a new SDK for a specific language, please follow the [Engine API Specification](../DOCUMENTATION.md#api-specification).
