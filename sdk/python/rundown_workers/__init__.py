from .client import Client
import functools
import sys
import signal
import threading
import time

# Global list of workers to start
_workers = []

def queue(name, host="http://localhost:8181", poll_interval=1.0, max_retries=3):
	def decorator(func):
		_workers.append((name, host, func, poll_interval, max_retries))
		
		@functools.wraps(func)
		def wrapper(payload):
			return func(payload)
		return wrapper
	return decorator

def enqueue(queue, payload, host="http://localhost:8181", timeout=None, max_retries=None, tag=None):
	"""Submit a job to a queue."""
	client = Client(host=host)
	return client.enqueue(queue, payload, timeout=timeout, max_retries=max_retries, tag=tag)

def run():
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
		queue_name, host, handler, interval, max_retries = worker_info
		client = Client(host=host)
		
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
