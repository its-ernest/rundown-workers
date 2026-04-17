# rundown-workers

Node.js SDK for [Rundown-Workers](https://github.com/its-ernest/rundown-workers) — a lightweight background job engine.

---

## Requirements

- Node.js 18+
- Rundown-Workers engine running (see [engine setup](https://github.com/its-ernest/rundown-workers#installation))

---

## Installation

```bash
npm install @rundownworker/rundown-workers
```

---

## Quick Start

### 1. Enqueue a job from your backend

```ts
import { enqueue } from "@rundownworker/rundown-workers";

await enqueue("email_sender", { to: "user@example.com" });
```

### 2. Create a worker process

**Functional style (simple)**

```ts
import { queue, run } from "@rundownworker/rundown-workers";

queue({ name: "email_sender" }, async (payload) => {
  await sendEmail(payload.to);
});

await run({ host: "http://localhost:8181" });
```

**Class style (more control)**

```ts
import { Worker } from "@rundownworker/rundown-workers";

const worker = new Worker({ host: "http://localhost:8181" });

worker.register({ name: "email_sender" }, async (payload) => {
  await sendEmail(payload.to);
});

await worker.start();
```

---

## API

### `enqueue(queue, payload, options?)`

Sends a job to the engine.

```ts
await enqueue(
  "image_processor",
  { url: "https://..." },
  {
    timeout: 60,
    max_retries: 3,
    host: "http://localhost:8181",
  }
);
```

| Option        | Type     | Default                 | Description                                  |
| ------------- | -------- | ----------------------- | -------------------------------------------- |
| `timeout`     | `number` | `300`                   | Max seconds the job is allowed to run        |
| `max_retries` | `number` | `0`                     | How many times the engine retries on failure |
| `host`        | `string` | `http://localhost:8181` | Engine URL                                   |

Returns the created job object or `null` if the engine is unreachable.

---

### `queue(config, handler)`

Registers a handler on the global worker instance. Use with `run()` for a minimal setup.

```ts
import { queue, run } from "@rundownworker/rundown-workers";

queue({ name: "email_sender" }, async (payload) => {
  await sendEmail(payload.to);
});

await run({ host: "http://localhost:8181" });
```

---

### `run(config?)`

Starts the global worker instance and begins polling all queues registered via `queue()`.

```ts
await run({
  host: "http://localhost:8181",
  pollInterval: 1000,
  concurrency: 2,
});
```

| Option         | Type       | Default                 | Description                                            |
| -------------- | ---------- | ----------------------- | ------------------------------------------------------ |
| `host`         | `string`   | `http://localhost:8181` | Engine URL                                             |
| `pollInterval` | `number`   | `2000`                  | How often (ms) to poll for new jobs                    |
| `concurrency`  | `number`   | `1`                     | Max parallel jobs across all queues                    |
| `onError`      | `function` | —                       | Called when a job fails, before the engine is notified |

---

### `new Worker(config)`

Creates a worker instance directly for more control.

```ts
const worker = new Worker({
  host: "http://localhost:8181",
  pollInterval: 1000,
  concurrency: 2,
  onError: (error, jobId, queue) => {
    console.error(`[${queue}] job ${jobId} failed:`, error.message);
  },
});
```

Accepts the same options as `run()`.

---

### `worker.register(config, handler)`

Registers a handler for a named queue. Chainable.

```ts
worker
  .register<{ to: string }>({ name: "email_sender" }, async (payload) => {
    await sendEmail(payload.to);
  })
  .register<{ orderId: string }>(
    { name: "webhook_dispatch", concurrency: 3 },
    async (payload) => {
      await dispatchWebhook(payload.orderId);
    }
  );
```

| Config        | Type     | Description                                        |
| ------------- | -------- | -------------------------------------------------- |
| `name`        | `string` | Queue name to poll                                 |
| `concurrency` | `number` | Max parallel jobs for this queue, overrides global |

---

### `worker.start()`

Starts polling all registered queues. Runs until the process is terminated.

```ts
await worker.start();
```

---

### `worker.stop()`

Stops polling and waits for all active jobs to finish before resolving.

```ts
await worker.stop();
```

Shutdown is also triggered automatically on `SIGINT` and `SIGTERM`.

---

## Timeout Behaviour

Timeout is set at enqueue time and enforced by the SDK at execution time.

```ts
await enqueue("slow_task", { file: "large.csv" }, { timeout: 30 });
```

If the handler does not resolve within 30 seconds, the SDK cancels it and reports failure to the engine. The engine then decides whether to retry based on `max_retries`.

> **Note:** Timeout enforcement works correctly for async handlers. Synchronous CPU-blocking handlers will not be interrupted until the event loop is free.

---

## Multiple Queues

A single worker process can handle multiple queues simultaneously.

```ts
const worker = new Worker({ host: "http://localhost:8181" });

worker
  .register({ name: "email_sender" }, handleEmail)
  .register({ name: "image_processor", concurrency: 4 }, handleImage)
  .register({ name: "webhook_dispatch" }, handleWebhook);

await worker.start();
```

Each queue runs its own independent poll loop.

---

## Error Handling

If a handler throws, the SDK:

1. Calls your `onError` hook (if provided)
2. Logs to `console.error`
3. Reports the failure to the engine via `POST /fail`
4. Continues processing other jobs — the worker never crashes

Network failures when reaching the engine are swallowed silently and polling resumes on the next interval.

---

## Typed Payloads

Use generics to type your payloads end to end.

```ts
type EmailJob = { to: string; subject: string };

worker.register<EmailJob>({ name: "email_sender" }, async (payload) => {
  // payload is EmailJob here
  await sendEmail(payload.to, payload.subject);
});
```

---

## Full Example

```ts
import { enqueue, Worker } from "@rundownworker/rundown-workers";

const worker = new Worker({
  host: "http://localhost:8181",
  pollInterval: 1000,
  concurrency: 2,
  onError: (error, jobId, queue) => {
    myLogger.error({ jobId, queue, error: error.message });
  },
});

worker
  .register<{ to: string; subject: string }>(
    { name: "email_sender" },
    async (payload) => {
      await sendEmail(payload.to, payload.subject);
    }
  )
  .register<{ orderId: string }>(
    { name: "receipt_generator", concurrency: 3 },
    async (payload) => {
      await generateReceipt(payload.orderId);
    }
  );

await worker.start();
```

Enqueue from anywhere in your app:

```ts
await enqueue(
  "email_sender",
  { to: "user@example.com", subject: "Your receipt" },
  {
    max_retries: 3,
    timeout: 10,
  }
);
```

---

## License

MIT
