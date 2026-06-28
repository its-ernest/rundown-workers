# SDK & Integration Guide

Because Rundown-Workers operates entirely over a clean HTTP JSON API, you can write workers in any language without any special dependencies.

---

## Writing a Custom Worker Poller

Implementing a worker in a language without an official SDK is straightforward. A worker simply runs an infinite loop carrying out the following steps:

1. **Poll**: Send a `POST /poll` request with the queue name and a batch limit.
2. **Execute**: For each job returned:
   - Run your processing logic inside a `try/except` (or `recover/catch`) wrapper.
   - Set a local timer matching the job's `timeout` value to abort if it hangs.
3. **Report**:
   - On success: Send `POST /complete` with the job `id`.
   - On exception: Send `POST /fail` with the job `id`.
4. **Throttle**: Sleep briefly if no jobs were returned (to avoid hammering the server).

### Python Native Mock Example

```python
import time
import requests

ENGINE_URL = "http://localhost:8181"
QUEUE = "video_transcoder"

def process_job(payload):
    print(f"Transcoding video: {payload['link']} to {payload['format']}")
    # Your worker logic goes here

def start_worker():
    print(f"Worker listening on queue: {QUEUE}")
    while True:
        try:
            # 1. Poll the engine
            resp = requests.post(f"{ENGINE_URL}/poll", json={"queue": QUEUE, "limit": 1})
            
            if resp.status_code == 204:
                time.sleep(2) # Queue is empty, rest a bit
                continue

            jobs = resp.json()
            for job in jobs:
                job_id = job["id"]
                payload = job["payload"] # Raw JSON payload structure
                
                try:
                    # 2. Execute
                    process_job(payload)
                    
                    # 3. Report Success
                    requests.post(f"{ENGINE_URL}/complete", json={"id": job_id})
                except Exception as e:
                    # 3. Report Failure
                    print(f"Job failed: {e}")
                    requests.post(f"{ENGINE_URL}/fail", json={"id": job_id})

        except Exception as conn_err:
            print(f"Connection error: {conn_err}")
            time.sleep(5)

if __name__ == "__main__":
    start_worker()
```

---

## SDK Status

We support the following experimental client packages:

- **[Python SDK](file:///run/media/arch/Extra/my/projects/its-ernest/rundown-workers/sdk/python/README.md)**
- **[Node.js SDK](file:///run/media/arch/Extra/my/projects/its-ernest/rundown-workers/sdk/nodejs/README.md)**
- **[Go SDK](file:///run/media/arch/Extra/my/projects/its-ernest/rundown-workers/sdk/go/README.md)**

*Note: Since the HTTP interface is extremely simple and stable, developers are highly encouraged to use standard HTTP libraries (`requests`, `fetch`, `net/http`) directly in their apps for robust long-term integration.*
