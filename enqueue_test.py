import rundown_workers as rw

rw.enqueue(queue="post_worker", payload="Hello from Rundown Script")
print("Job enqueued successfully.")
