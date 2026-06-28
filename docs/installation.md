# Installation Guide

Rundown-Workers can be run either via Docker (highly recommended) or natively as a binary on your host machine.

---

## 1. Running with Docker Compose (Recommended)

To run the engine alongside your project services, add the official Docker Hub image `itsernest/rundown-workers:latest` to your `compose.yml`.

### Example Compose Configuration

```yaml
services:
  # Rundown Workers Background Engine
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

Once running, any container in the same Docker network can communicate with the engine at `http://rundown:8181`.

---

## 2. Running on the Host Machine

You can build Rundown-Workers from source and run it natively.

### Prerequisites
- Go 1.25.6 or later
- [Taskfile](https://taskfile.dev) installed (optional, but recommended for build shortcuts)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/its-ernest/rundown-workers.git
cd rundown-workers

# Build the binary using Taskfile
task build
```

This creates the compiled binary in `./bin/rundown-engine`.

### Running with SQLite (Default)
By default, the engine starts up with a SQLite storage backend:

```bash
./bin/rundown-engine
```

### Running with PostgreSQL
To connect the engine to an existing PostgreSQL database on your host machine:

```bash
export DATABASE_URL="postgres://flare_admin:super_secure_db_password_123@localhost:5432/flare_boost_db?sslmode=disable"
./bin/rundown-engine --store=postgres
```

---

## Configuration Reference

The engine can be configured using environment variables or CLI flags:

| Environment Variable | CLI Flag | Description | Default |
|----------------------|----------|-------------|---------|
| `RUNDOWN_STORE`      | `--store`| Storage backend: `sqlite` or `postgres` | `sqlite` |
| `DATABASE_URL`       | —        | PostgreSQL connection string (Required if store is `postgres`) | — |
| `RUNDOWN_DB_PATH`    | —        | Path to SQLite file database (Used when store is `sqlite`) | `rundown_v2.db` |
| `RUNDOWN_HOST`       | `--host` | Host address to bind the HTTP server to | `0.0.0.0` |
| `RUNDOWN_PORT` / `PORT` | `--port` | Port to bind the HTTP server to | `8181` |
