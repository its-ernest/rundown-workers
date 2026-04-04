# Rundown-Workers

### Latest Release: v0.1.0 (Stable)

The Rundown-Workers Engine is now even more flexible with custom port support and pre-compiled binaries for Windows and Linux (including ARM64).

Rundown-Workers is a lightweight, language-agnostic workflow executor for developers who need reliable background job processing without heavy infrastructure.

It combines a Go-based core engine with simple SDKs (Python, Node.js, etc.), allowing tasks to be defined and executed in any language.

---

## Philosophy

Most workflow systems are powerful but unnecessarily complex.

Rundown-Workers follows a simpler principle:

Keep the engine minimal and let execution happen in the language where the code already lives.

* The engine orchestrates
* The SDK executes

---

## Architecture Overview

The system consists of:

* A Go engine (HTTP API + SQLite)
* Language-specific SDKs (Python, Node.js, Go, etc.)
* Workers that poll and execute jobs

Communication happens over HTTP, making the system language-agnostic.

---

## How It Works

1. A job is enqueued into the engine
2. A worker polls the engine for jobs
3. The engine assigns a job
4. The SDK executes the job locally
5. The worker reports completion

---

## Installation

### 1. Install Rundown Workers
You must have it installed first, before you can use the SDKs in your project(such as Python package, Go package, Node.js package, etc).

```bash
# if you want to use pre-built binary
# download from releases

# for linux
# run this from your project root
$ curl -L https://github.com/its-ernest/rundown-workers/releases/download/v0.1.0/engine -o rundown-workers/engine

# for windows
# run this from your project root
$ curl -L https://github.com/its-ernest/rundown-workers/releases/download/v0.1.0/engine.exe -o rundown-workers/engine.exe
```

```bash
# if you want manual build
$ git clone https://github.com/its-ernest/rundown-workers.git
$ cd rundown-workers
```

### 2. Run the engine

```bash
# if you don't have go installed
# download rundown-workers binary from releases
# and run this command

$ ./rundown-workers/engine # on linux
$ ./rundown-workers/engine.exe # on windows
```

```bash
# manual builds
# if you have go installed
$ git clone https://github.com/yourusername/rundown-workers.git
$ cd rundown-workers
$ go run cmd/worker/main.go
```

The server starts at:

```bash
http://localhost:8181

# if you want to change port
$ ./rundown-workers/engine --port 8080
```

---

## SDKs

You can schedule and manage worker jobs in your backend using the SDKs. For instance, if your backend is in Python, you can use the Python SDK to schedule and manage worker jobs.

- [Python](sdk/python/README.md)
- [Node.js](sdk/nodejs/README.md)
- [Go](sdk/go/README.md)

# Use cURL as fallbackc for now if your backend is not in the SDK list above

```bash
# Enqueue a job
curl -X POST http://localhost:8181/enqueue \
-H "Content-Type: application/json" \
-d '{
  "queue": "post_worker",
  "payload": "Hello from Rundown"
}'

# Poll for a job
curl -X POST http://localhost:8181/poll \
-H "Content-Type: application/json" \
-d '{
  "queue": "post_worker"
}'

# Mark job as complete
curl -X POST http://localhost:8181/complete \
-H "Content-Type: application/json" \
-d '{
  "id": "job-id"
}'

# Mark job as failed
curl -X POST http://localhost:8181/fail \
-H "Content-Type: application/json" \
-d '{
  "id": "job-id"
}'
```

## Basic Usage (example in Python sdk)

### Step 1 — Define a Worker (Python)

```python
import rundown_workers as rw

# This task will fail if it takes longer than 2 seconds
rw.enqueue(queue="greetings", payload="Hello!", timeout=2)

# This task will retry 3 times if it fails
rw.enqueue(queue="greetings", payload="Hello!", max_retries=3)

# This task will retry 3 times if it fails and will time out after 2 seconds
rw.enqueue(queue="greetings", payload="Hello!", timeout=2, max_retries=3)

# This actively fetches and executes jobs
@rw.queue(name="greetings", host="http://localhost:8181")
def run_work(payload):
    print("Processing:", payload)
    return True
```

Run the worker:

```bash
python worker.py
```

This starts a background process that continuously polls for jobs.

---

### Step 2 — Enqueue a Job

```bash
curl -X POST http://localhost:8181/enqueue \
-H "Content-Type: application/json" \
-d '{
  "queue": "post_worker",
  "payload": "Hello from Rundown"
}'
```

---

## What Happens Next

* The engine stores the job in SQLite
* The worker polls the /poll endpoint
* A job is assigned
* The function executes
* The worker calls /complete
* The job is marked as done

---

## Core Concepts

### Queue

A named channel for jobs.

Examples:

```
post_worker
email_sender
image_processor
```

---

### Job

A unit of work:

```
{
  "id": "uuid",
  "queue": "post_worker",
  "payload": "data",
  "status": "pending"
}
```

---

### Worker

A process that:

* Polls the engine
* Executes jobs
* Reports completion

---

## Job Lifecycle

```
pending -> running -> done
```

---

## Why This Design

Rundown-Workers avoids forcing all logic into a single language.

It allows:

* Python for data processing
* Node.js for asynchronous tasks
* Go for performance-sensitive work

All coordinated through a single lightweight engine.

---

## Current Limitations

This project is in an early stage.

Missing features include:

* Retry mechanism
* Job timeouts
* Dead letter queue
* Scheduling or delayed jobs
* Authentication
* Observability (logging and metrics)

---

## Roadmap

* Retry and backoff strategy
* Timeout handling and job recovery
* Multiple workers per queue
* SDK for Node.js
* CLI tooling
* Monitoring dashboard
* Optional distributed mode (e.g. PostgreSQL)

---

## Contributing

Contributions are welcome.

Areas to start:

* SDK improvements
* Retry logic
* Testing

---

## License

MIT

---

## Idea behind this project

Rundown-Workers is not designed to compete with complex workflow engines.

It is built for simplicity, clarity, and control.

Simple systems scale better because they are easier to reason about and harder to break.
