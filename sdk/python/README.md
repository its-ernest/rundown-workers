# Rundown-Workers Python SDK (v0.1.0)

[![Go Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen)](https://github.com/its-ernest/rundown-workers)

A highly developer-friendly, lightweight Python SDK for implementing and running workflow workers.

## Installation

To install the SDK locally in editable mode (during development):

```bash
# install this once
pip install rundown-workers

#or
# From the project root (Manual build)
pip install -e sdk/python
```

## Basic Usage

The SDK uses decorators to register worker functions.

```python
import rundown_workers as rw

# Register a simple task
@rw.queue(name="greetings", max_retries=3)
def hello_task(payload):
    print(f"[*] Received: {payload}")
    return True

if __name__ == "__main__":
    # Start all registered workers in separate polling threads
    rw.run()
```

### Advanced Features

#### 1. Retries with Exponential Backoff
When enqueuing a job or defining a queue, you can specify `max_retries`. If your function raises an exception, the engine will automatically reschedule the job with an increasing delay (e.g., 5s, 20s, 45s).

#### 2. Local Timeout Enforcement
You can set a `timeout` (in seconds) for each task. If your function hangs beyond this period, the worker will automatically:
1. Detect the timeout.
2. Report the failure to the engine.
3. Move on to the next available job.

```python
# This task will fail if it takes longer than 2 seconds
rw.enqueue(queue="greetings", payload="Hello!", timeout=2)
```

## Internal Architecture

The Python SDK uses `threading` to handle local polling for multiple queues simultaneously and to monitor task execution times.

- **Polling Loop**: Each queue has its own daemon thread continuously hitting the `/poll` endpoint.
- **Task Execution**: Handlers are executed in a temporary thread with a `join(timeout=...)` call to ensure the worker process doesn't hang.
