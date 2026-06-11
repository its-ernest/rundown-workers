from .client import Client
import functools
import sys
import signal
import threading
import time

import os

# Default Rundown Engine URL from environment
DEFAULT_URL = os.getenv("RUNDOWN_URL", os.getenv("RUNDOWN_ENGINE_URL", "http://localhost:8181"))

def queue(name, url=None, poll_interval=1.0, max_retries=3):
	def decorator(func):
		_workers.append((name, url or DEFAULT_URL, func, poll_interval, max_retries))
		
		@functools.wraps(func)
		def wrapper(payload):
			return func(payload)
		return wrapper
	return decorator

def enqueue(queue, payload, url=None, timeout=None, max_retries=None, tag=None):
	"""Submit a job to a queue."""
	client = Client(url=url or DEFAULT_URL)
	return client.enqueue(queue, payload, timeout=timeout, max_retries=max_retries, tag=tag)

def run(url=None):
	"""Main entry point to start all decorated workers."""
	if not _workers:
		print("[!] No workers registered. Use @rw.queue to register functions.")
		return

	# Handle graceful shutdown (Ctrl+C)
	def signal_handler(sig, frame):
		print("\n[*] Rundown-Workers worker stopped.")
		sys.exit(0)
	signal.signal(signal.SIGINT, signal_handler)

	# Start each worker in a separate polling thread
	threads = []
	for worker_info in _workers:
		queue_name, worker_url, handler, interval, max_retries = worker_info
		# Override if url is provided to run()
		final_url = url or worker_url
		client = Client(url=final_url)
		
		t = threading.Thread(
			target=client.start_worker, 
			args=(queue_name, handler, interval),
			daemon=True
		)
		t.start()
		threads.append(t)
	
	# Keep main thread alive
	while True:
		time.sleep(1)
