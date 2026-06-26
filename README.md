# Rundown-Workers

Rundown-Workers is a lightweight, language-agnostic background job and workflow executor. It manages queue states, transaction timeouts, and retries, while allowing your code (workers) to execute anywhere in any language over simple HTTP JSON APIs.

---

## Docker Compose Quick Start

To add the engine to your Docker Compose setup, reference the official Docker Hub image `ernesd5050/rundown-workers:latest`. 

Here is the common configuration you drop directly into your `compose.yml`:

```yaml
services:
  # RUNDOWN WORKERS BACKGROUND ENGINE
  rundown:
    image: ernesd5050/rundown-workers:latest
    container_name: rundown-workers-engine
    ports:
      - "8181:8181"
    environment:
      - RUNDOWN_STORE=postgres
      - DATABASE_URL=postgres://flare_admin:super_secure_db_password_123@flare-postgres-db:5432/flare_boost_db?sslmode=disable
    restart: unless-stopped
```

Once declared in your `compose.yml`, any other container in the same service network can poll or dispatch jobs from the engine using its service link:
`http://rundown:8181`

---

## Host Machine Quick Start (Binary)

If you prefer to run the engine directly on your host machine without Docker:

### 1. Download & Build from Source
```bash
git clone https://github.com/its-ernest/rundown-workers.git
cd rundown-workers
task build
```

### 2. Run with SQLite (Default)
```bash
./bin/rundown-engine
```

### 3. Run with PostgreSQL
```bash
export DATABASE_URL="postgres://flare_admin:super_secure_db_password_123@localhost:5432/flare_boost_db?sslmode=disable"
./bin/rundown-engine --store=postgres
```

---

## Configuration Reference

Customize the behavior of the engine via environment variables or CLI flags:

| Environment Variable | CLI Flag | Description | Default |
|----------------------|----------|-------------|---------|
| `RUNDOWN_STORE`      | `--store`| Storage backend to use: `sqlite` or `postgres` | `sqlite` |
| `DATABASE_URL`       | —        | DSN Connection string (Required for `postgres`) | — |
| `RUNDOWN_DB_PATH`    | —        | SQLite file database path (Used when store is `sqlite`) | `rundown_v2.db` |
| `RUNDOWN_HOST`       | `--host` | Network host to bind server to | `0.0.0.0` |
| `RUNDOWN_PORT` / `PORT` | `--port` | Network port to bind server to | `8181` |

---

## Local Development (Taskfile Commands)

For local modifications and engine development, we use [Taskfile](https://taskfile.dev):

```bash
# Build binary for current platform
task build

# Build binaries for all supported platforms
task build:all

# Start local dev stack with docker-compose override
task docker:up

# Stop docker-compose stack
task docker:down

# Run unit tests
task test

# Clean build artifacts
task clean
```

---

## HTTP API Reference

### 1. Enqueue a Job
```bash
curl -X POST http://localhost:8181/enqueue \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "video_transcoder",
    "tag": "user_upload_123",
    "payload": {
      "link": "https://example.com/video.mp4",
      "format": "mp4"
    },
    "timeout": 360,
    "max_retries": 3
  }'
```

### 2. Poll for Jobs
```bash
curl -X POST http://localhost:8181/poll \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "video_transcoder",
    "limit": 5
  }'
```

### 3. Mark Job as Completed
```bash
curl -X POST http://localhost:8181/complete \
  -H "Content-Type: application/json" \
  -d '{"id": "<job-uuid>"}'
```

### 4. Mark Job as Failed
```bash
curl -X POST http://localhost:8181/fail \
  -H "Content-Type: application/json" \
  -d '{"id": "<job-uuid>"}'
```

### 5. Get Job Status by ID
```bash
curl http://localhost:8181/status/<job-uuid>
```

### 6. Get Job Details by Tag
```bash
curl http://localhost:8181/details/<tag>
```

### 7. Get Technical Documentation (HTML)
```bash
curl http://localhost:8181/docs
```

---

## License
Distributed under the **MIT License**.