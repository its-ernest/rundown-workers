# Rundown-Workers

Rundown-Workers is a lightweight, language-agnostic workflow executor for developers who need reliable background job processing without heavy infrastructure.

It combines a Go-based core engine with structured HTTP primitives, allowing tasks to be defined, batched, and executed cleanly in any language.

---

## Philosophy

Most workflow systems are powerful but unnecessarily complex.

Rundown-Workers follows a simpler principle: **Keep the engine minimal and let execution happen in the language where the code already lives.**

* The engine orchestrates state.
* Your worker code executes the logic.

---

## Architecture Overview

The system consists of:

* **A Go Engine:** A high-performance HTTP API backed by a local SQLite store (`rundown_v2.db`).
* **Interval/Continuous Workers:** External scripts or daemons that fetch task items via atomic batch calls, parse payloads, and handle execution safely.

Communication happens purely over native HTTP JSON payloads, making the entire ecosystem inherently language-agnostic.

---

## How It Works

1. A job with a structured JSON payload is enqueued into the engine.
2. A worker polls the `/poll` endpoint (individually or in batches up to a specified limit).
3. The engine updates the row transaction parameters and assigns unique tracking IDs.
4. The worker executes the execution code locally with isolated recovery guards.
5. The worker posts a validation confirmation back to mark completion or track execution failure.

---

## Installation

### 1. Install Rundown Workers

You must have the engine running before you can dispatch tasks or trigger background execution nodes.

```bash
# Download a pre-built binary from releases

# For Linux (replace amd64 with your architecture)
$ curl -L https://github.com/its-ernest/rundown-workers/releases/download/v0.2.0/engine-linux-amd64 -o rundown-workers/engine

# For Windows
$ curl -L https://github.com/its-ernest/rundown-workers/releases/download/v0.2.0/engine-windows-amd64.exe -o rundown-workers/engine.exe

```

```bash
# Or build manually from source
$ git clone https://github.com/its-ernest/rundown-workers.git
$ cd rundown-workers
$ make build

```

### 2. Run the Engine

```bash
# Run pre-built engine binary
$ ./rundown-workers/engine       # Linux
$ ./rundown-workers/engine.exe   # Windows

```

```bash
# Run manual source build
$ go run main.go

```

The server binds and initializes by default at: `http://localhost:8181`

```bash
# Optional: Change the binding port and host tracking fields
$ ./rundown-workers/engine --host 0.0.0.0 --port 8080

```

---

## SDKs & HTTP Core API

**NOTE: SDKS AREN'T ALWAYS UP-TO-DATE. IT IS STRONGLY RECOMMENDED TO STICK TO NATIVE NATIVE HTTP / cURL CALLS FOR OPERATING**

You can write direct API wrappers in your backend using the structural links below, or use raw HTTP fallback channels for stable processing:

* [Python](sdk/python/README.md)
* [Node.js](sdk/nodejs/README.md)
* [Go](sdk/go/README.md)

### Native Core HTTP Operations

#### 1. Enqueue a Job (Accepts Native JSON Object Payloads)

```bash
curl -X POST http://localhost:8181/enqueue \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "post_worker",
    "payload": {
      "video_link": "https://www.tiktok.com/@growth_ops/video/7341",
      "type": "organic"
    },
    "timeout": 360,
    "max_retries": 3
  }'

```

#### 2. Batch Poll for Jobs (Using Limit Constraints)

```bash
curl -X POST http://localhost:8181/poll \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "post_worker",
    "limit": 10
  }'

```

*Returns a JSON array block listing up to 10 unique job objects with assigned tracking IDs.*

#### 3. Mark Job as Complete

```bash
curl -X POST http://localhost:8181/complete \
  -H "Content-Type: application/json" \
  -d '{
    "id": "abc-123-unique-job-uuid"
  }'

```

#### 4. Mark Job as Failed

```bash
curl -X POST http://localhost:8181/fail \
  -H "Content-Type: application/json" \
  -d '{
    "id": "abc-123-unique-job-uuid"
  }'

```

---

## Core Concepts

### Queue

A named pipeline channel grouping specific varieties of task processing workloads together.

* *Examples:* `post_worker`, `email_sender`, `image_processor`

### Job

An atomic, uniquely identified unit of persistent work containing metadata configurations and a structural JSON payload data segment.

### Worker

An isolated application or background daemon process that queries execution batches at designated time frequencies, handles execution safely via runtime panics/recovery mechanisms, and reports structural statuses back to the engine.

---

## Job Lifecycle

```
[ pending ] ──( Poll / Batch Limit )──> [ running ] ──┬──( Success )──> [ done ]
                                                      └──( Failure )──> [ failed ]

```

---

## Why This Design

Rundown-Workers avoids forcing all your core system infrastructure dependencies into a single runtime ecosystem. It lets you write high-performance orchestration layers where they make the most sense:

* **Python** for machine learning pipelines and quick automation tracking hooks.
* **Node.js** for high-throughput asynchronous file streaming logic.
* **Go** for blistering-fast backend systems, resource concurrency management, and API architectures.

All seamlessly coordinated through a decoupled, atomic SQLite core transaction broker.

---

## Current Limitations & Roadmap

This system is actively evolving. Current layout constraints and future roadmap milestones include:

* [x] Structured JSON Object Payload Support
* [x] Array-based Multi-Job Batch Polling (`limit` parameter flags)
* [x] Automatic 30s Database Staleness Cleanup and Job Recovery Ticker
* [ ] Dead Letter Queue (DLQ) Isolation Arrays
* [ ] Multi-node Authentication Layering
* [ ] Native Prom/Grafana Infrastructure Metrics Monitoring Dashboard

---

## Contributing & License

Contributions are welcome. Areas to start focus on include SDK updates, automated integration test coverage parameters, and custom CLI orchestration tools.

Distributed under the **MIT License**.

---

### Philosophy Note

Rundown-Workers isn't built to compete with hyper-complex, heavily bloated microservice workflow engines. It is engineered purposefully for simplicity, clarity, and total control. Simple code paths scale cleaner because they are inherently predictable and incredibly hard to break.