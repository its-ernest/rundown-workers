import rundown_workers as rw

print("Enqueuing failing task (should be retried twice)...")
rw.enqueue(queue="retry_queue", payload="fail-me", max_retries=2)

print("Enqueuing slow task (should timeout after 2s)...")
rw.enqueue(queue="timeout_queue", payload="slow-me", timeout=2)

print("Done.")
