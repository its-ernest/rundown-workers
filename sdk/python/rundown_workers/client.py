import requests
import time
import threading

class Client:
	def __init__(self, url="http://localhost:8181"):
		self.url = url.rstrip("/")

	def enqueue(self, queue, payload, timeout=None, max_retries=None, tag=None):
		"""
		Submits a new task to the Rundown-Workers engine.

		:param queue: The name of the queue to join.
		:param payload: The data content of the job.
		:param timeout: Optional override for the execution time limit (in seconds).
		:param max_retries: Optional override for the maximum retry count.
		:param tag: Optional tag for the job.
		"""
		url = f"{self.url}/enqueue"
		data = {"queue": queue, "payload": payload}
		if tag is not None:
			data["tag"] = tag
		if timeout:
			data["timeout"] = timeout
		if max_retries is not None:
			data["max_retries"] = max_retries
		resp = requests.post(url, json=data)
		resp.raise_for_status()
		return resp.json()

	def poll(self, queue):
		url = f"{self.url}/poll"
		data = {"queue": queue}
		resp = requests.post(url, json=data)
		if resp.status_code == 204:
			return None
		resp.raise_for_status()
		return resp.json()

	def complete(self, job_id):
		url = f"{self.url}/complete"
		data = {"id": job_id}
		resp = requests.post(url, json=data)
		resp.raise_for_status()

	def fail(self, job_id):
		url = f"{self.url}/fail"
		data = {"id": job_id}
		resp = requests.post(url, json=data)
		resp.raise_for_status()

	def start_worker(self, queue_name, handler, poll_interval=1.0):
		"""
		Starts a continuous polling loop for a specific queue.

		This method supports thread-safe local execution and will automatically
		fail jobs if they exceed their timeout limit.
		"""
		print(f"[*] Starting worker for queue '{queue_name}'...")
		while True:
			try:
				job = self.poll(queue_name)
				if not job:
					time.sleep(poll_interval)
					continue

				print(f"[*] Job {job['id']} assigned. Executing...")
				
				# Get timeout from job (default 300s)
				job_timeout = job.get('timeout', 300)
				
				# Container for thread results
				# [success, error_message]
				state = [False, ""]
				
				def run_handler():
					try:
						handler(job['payload'])
						state[0] = True
					except Exception as e:
						state[1] = str(e)

				task_thread = threading.Thread(target=run_handler, daemon=True)
				task_thread.start()
				task_thread.join(timeout=job_timeout)

				if task_thread.is_alive():
					print(f"[!] Job {job['id']} timed out after {job_timeout}s.")
					try:
						self.fail(job['id'])
					except Exception as fe:
						print(f"[!] Error reporting timeout: {fe}")
				elif state[0]:
					try:
						self.complete(job['id'])
						print(f"[*] Job {job['id']} completed successfully.")
					except Exception as ce:
						print(f"[!] Error reporting completion: {ce}")
				else:
					print(f"[!] Job {job['id']} failed: {state[1]}")
					try:
						self.fail(job['id'])
					except Exception as fe:
						print(f"[!] Error reporting failure: {fe}")

			except Exception as e:
				print(f"[!] Worker polling error: {e}")
				time.sleep(5)
