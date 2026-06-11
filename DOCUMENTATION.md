# Rundown-Workers Technical Documentation

Rundown-Workers is a language-agnostic background task executor designed for simplicity, reliability, and local execution.

## Core Architecture

The system follows an **Engine-Worker** pattern:

*   **Engine (Go)**: A stateless HTTP API backed by SQLite that manages job state, retries, and persistence.
*   **SDK (Python/Node)**: A language-specific library that tasks are registered in.
*   **Worker**: A long-running process that polls the engine and executes registered functions.

## Feature Guide

### 1. Reliable Retries
Jobs can be enqueued with a `max_retries` value. 
- If a task raises an exception, the Engine will automatically reschedule it.
- **Exponential Backoff**: Retries are scheduled using the formula `5 * (retry_count^2)` seconds in the future.

### 2. High-Concurrency Polling
To prevent multiple workers from pulling the same task, the engine uses **IMMEDIATE transactions** in SQLite. This ensures atomic "claim and mark" operations even under heavy load.

### 3. Timeout Enforcement
Every job has a `timeout` period (default 300s).
- **SDK Enforcement**: The worker wraps your function in a thread and will report a failure if it hangs.
- **Engine Recovery**: The "Staleness Checker" in the engine periodically moves jobs that have "lost" their workers back to the pending queue.

## API Specification

### `POST /enqueue`
Add a job to a named queue.

**Payload:**
```json
{
  "queue": "images",
  "payload": "path/to/image.jpg",
  "timeout": 60,
  "max_retries": 5
}
```

### `POST /poll`
Workers call this to request their next task.

### `POST /complete`
Mark a job as successfully finished.

### `POST /fail`
Manually report that a task failed and should be retried or marked as permanently failed.

## Python SDK Usage

```python
import rundown_workers as rw

# Register a worker
@rw.queue(name="my_queue", max_retries=5)
def handle_task(payload):
    print("Doing work:", payload)

if __name__ == "__main__":
    rw.run()
```
