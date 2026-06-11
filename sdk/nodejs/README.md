# Rundown-Workers Node.js SDK (Planned)

A future implementation of the Rundown-Workers SDK for JavaScript/TypeScript.

## Implementation Blueprint

When building the Node.js SDK, follow these design goals:

### Interface Goal (JavaScript)

```javascript
import rw from 'rundown-workers';

// Register a worker function
rw.queue({ name: 'email-sender', maxRetries: 5 }, async (payload) => {
    console.log('Sending email:', payload);
    await sendMail(payload);
});

// Start workers
rw.run();
```

### Key Considerations

1.  **Async/Await**: The SDK must support asynchronous handlers and properly await them.
2.  **Abortion/Timeout**: Use `AbortController` or a custom Promise timer to enforce the task `timeout`.
3.  **Concurrency**: In Node.js, multiple queues can be polled within the same event loop without separate threads, but each `rw.run()` should manage its own intervals.
4.  **Error Handling**: Automatically catch unhandled rejection/exceptions within the handler and report to `/fail`.

---

## Status: Planning

We're looking for contributors to help build out the Node.js SDK based on the [Engine API Specification](../../DOCUMENTATION.md#api-specification).
