# Usage & API Guide

Rundown-Workers orchestrates tasks using simple HTTP JSON payloads. This page details the core concepts, lifecycle stages, and the complete HTTP API reference.

---

## Core Concepts

- **Queue**: A named pipeline (e.g. `video_transcoder`, `email_sender`) grouping similar types of task workloads together.
- **Job**: An atomic unit of persistent work containing metadata (timeouts, retries, optional tags) and a structural JSON payload.
- **Worker**: Any script or process running in any language that polls the engine for jobs, executes the logic, and reports the results back.

---

## Job Lifecycle

A job moves through the following states:

```
[ pending ] ──( Poll / Batch Limit )──> [ running ] ──┬──( Success )──> [ done ]
                                                      └──( Failure )──> [ failed ]
```

1. **Pending**: Enqueued and waiting to be claimed by a worker.
2. **Running**: Claimed by a worker. The worker has a window defined by the job's `timeout` (seconds) to finish.
3. **Done**: Marked completed successfully.
4. **Failed**: Unsuccessful after exhaustion of `max_retries`.

### Timeout & Staleness Recovery
Every 30 seconds, the engine runs an internal staleness recovery loop. If a job remains in the `running` state longer than its specified `timeout` limit (due to worker crashing, network loss, etc.), the engine automatically:
- Increments the job's retry count.
- Resets its status back to `pending` so other workers can pick it up.

---

## HTTP API Reference

### 1. Enqueue a Job (`POST /enqueue`)
Adds a job to a named queue. The `payload` accepts arbitrary JSON structures.

**Request:**
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

**Parameters:**
- `queue` (string, required): The target queue name.
- `tag` (string, optional): A queryable label to easily track related jobs.
- `payload` (JSON object/array/value, required): The task arguments.
- `timeout` (int, optional): Max execution time in seconds before it is marked stale. Defaults to `300`.
- `max_retries` (int, optional): Max retry attempts before permanently failing. Defaults to `0`.

---

### 2. Poll for Jobs (`POST /poll`)
Claims one or more jobs from a queue. The engine uses atomic lock/claim updates to ensure high-concurrency safety.

**Request:**
```bash
curl -X POST http://localhost:8181/poll \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "video_transcoder",
    "limit": 5
  }'
```

**Parameters:**
- `queue` (string, required): Queue to pull from.
- `limit` (int, optional): Maximum batch size of jobs to pull. Defaults to `1`.

**Response:**
Returns a JSON array of job objects, or `204 No Content` if the queue is empty.

---

### 3. Complete a Job (`POST /complete`)
Marks a job as successfully finished.

**Request:**
```bash
curl -X POST http://localhost:8181/complete \
  -H "Content-Type: application/json" \
  -d '{"id": "c7a9df7b-1188-46ab-bb1e-a4c3f58e23f9"}'
```

---

### 4. Fail a Job (`POST /fail`)
Reports a job execution failure. If the job has retries remaining, it will be scheduled for a retry in the future using **exponential backoff** (`5 * retry_count^2` seconds). Otherwise, it is marked as `failed`.

**Request:**
```bash
curl -X POST http://localhost:8181/fail \
  -H "Content-Type: application/json" \
  -d '{"id": "c7a9df7b-1188-46ab-bb1e-a4c3f58e23f9"}'
```

---

### 5. Get Job Status (`GET /status/:id`)
Retrieves the full details and current status of a job by its unique UUID.

**Request:**
```bash
curl http://localhost:8181/status/c7a9df7b-1188-46ab-bb1e-a4c3f58e23f9
```

---

### 6. Get Job Details by Tag (`GET /details/:tag`)
Retrieves the most recently created job details matching a specific tag string.

**Request:**
```bash
curl http://localhost:8181/details/user_upload_123
```

---

### 7. Interactive Documentation (`GET /docs`)
Serves the rendered project documentation as an HTML page directly from the engine.

**Request:**
```bash
curl http://localhost:8181/docs
```
