# Rundown-Workers Go SDK (v0.1.0)

A lightweight, concurrent Go SDK for building workers that target the Rundown-Workers engine.

## Installation

```go
import "github.com/its-ernest/rundown-workers/sdk/go"
```

## Basic Usage

The Go SDK allows you to register multiple workers across different queues and run them in a single process.

```go
package main

import (
    "fmt"
    "time"
    rw "github.com/its-ernest/rundown-workers/sdk/go"
)

func main() {
    // Register a worker for the "emails" queue
    rw.Queue("emails", func(payload string) bool {
        fmt.Println("Processing email payload:", payload)
        return true // return true for success (complete), false for failure (retry)
    }, 1 * time.Second)

    // Start all workers and block until interrupted (SIGINT/SIGTERM)
    rw.Run("http://localhost:8181")
}
```

## Key Features

### 1. High Concurrency
Unlike the Python SDK, which uses system threads, the Go SDK leverages **Goroutines**. You can register hundreds of queues in a single binary with minimal resource overhead.

### 2. Local Timeout Enforcement
Every job specifies a `timeout`. The SDK uses an internal timer to ensure your handler doesn't hang. If it exceeds the timeout, the job is automatically reported as failed to the engine.

### 3. Graceful Shutdown
When you call `rw.Run()`, it automatically listens for system signals (like Ctrl+C). It will wait for any active job handlers to finish before exiting, ensuring no "lost" jobs in your local worker.
