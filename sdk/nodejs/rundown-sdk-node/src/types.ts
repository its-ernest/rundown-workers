export type Handler<T = unknown> = (payload: T) => Promise<void> | void;

export interface EnqueueOptions {
  timeout?: number;
  max_retries?: number;
  tag?: string;
  host?: string;
}

export interface QueueConfig {
  name: string;
  maxRetries?: number;
  concurrency?: number;
}

export interface WorkerConfig {
  host: string;
  pollInterval?: number;
  concurrency?: number;
  onError?: (error: Error, jobId: string, queue: string) => void;
}

export interface Job<T = unknown> {
  id: string;
  queue: string;
  tag?: string;
  payload: T;
  retry_count?: number;
  timeout?: number;
}

export interface EnqueueResult {
  id: string;
}

export interface QueueRegistration<T = unknown> {
  config: QueueConfig;
  handler: Handler<T>;
}
