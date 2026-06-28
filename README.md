# Rundown-Workers

Rundown-Workers is a lightweight, language-agnostic background task and workflow executor. It manages queue states, transaction timeouts, and retries, while allowing your code (workers) to execute anywhere in any language over simple HTTP JSON APIs.

---

## Quick Links

- **[Installation Guide](docs/installation.md)** — Docker Compose and host machine installation.
- **[Configuration & Usage](docs/usage.md)** — HTTP API reference, timeout settings, and lifecycle rules.
- **[SDKs & Custom Workers](docs/sdk.md)** — Official SDKs and implementing custom workers.

---

## Docker Compose Snippet

Reference the official Docker Hub image `itsernest/rundown-workers:latest` directly in your `compose.yml` to provision background job scheduling:

```yaml
services:
  # RUNDOWN WORKERS BACKGROUND ENGINE
  rundown:
    image: itsernest/rundown-workers:latest
    container_name: rundown-workers-engine
    ports:
      - "8181:8181"
    environment:
      - RUNDOWN_STORE=postgres
      - DATABASE_URL=postgres://flare_admin:super_secure_db_password_123@flare-postgres-db:5432/flare_boost_db?sslmode=disable
    restart: unless-stopped
```

Any container in the same service network can poll or dispatch jobs from the engine at `http://rundown:8181`.

---

## Features

- **Language Agnostic**: Workers can poll and process tasks using simple JSON HTTP requests.
- **Dual Engine Backends**: Choose between embedded SQLite for zero-config setups or PostgreSQL for production-ready, scaled compose architectures.
- **Concurrency Safe**: Uses transactional database locking to ensure jobs are claimed exactly once, even under heavy parallel polling.
- **Failures & Automatic Backoff**: Scheduled retries calculate automatic exponential backoffs.
- **Staleness Loop**: Auto-recovers jobs whose workers crashed or hung.

---

## License
Distributed under the **MIT License**.