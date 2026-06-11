import rundown_workers as rw
import time

@rw.queue(name="retry_queue", max_retries=2)
def failing_task(payload):
    print(f"[*] Executing failing task with: {payload}")
    raise Exception("Boom! Task failed.")

@rw.queue(name="timeout_queue")
def slow_task(payload):
    print(f"[*] Executing slow task with: {payload}")
    time.sleep(10) # Longer than the 2s timeout we'll set
    print("[*] Slow task finished (should not have completed if timed out)")

if __name__ == "__main__":
    rw.run()
