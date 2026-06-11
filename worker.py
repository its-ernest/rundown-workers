import rundown_workers as rw

@rw.queue(name="post_worker", host="http://localhost:8181")
def run_work(payload):
    print(f"[*] Processing job: {payload}")

if __name__ == "__main__":
    # Start the polling loop
    rw.run()
