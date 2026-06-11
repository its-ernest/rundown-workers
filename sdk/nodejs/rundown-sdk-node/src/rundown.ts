import { Worker } from "./worker.js";
import type { Handler, QueueConfig, WorkerConfig } from "./types.js";

let instance: Worker | null = null;

const DEFAULT_URL = process.env.RUNDOWN_URL || process.env.RUNDOWN_ENGINE_URL || "http://localhost:8181";

function getInstance(config?: WorkerConfig): Worker {
  if (!instance) {
    instance = new Worker(config ?? { host: DEFAULT_URL });
  }
  return instance;
}

export function queue<T = unknown>(
  config: QueueConfig,
  handler: Handler<T>
): void {
  getInstance().register(config, handler);
}

export async function run(config?: WorkerConfig): Promise<void> {
  await getInstance(config).start();
}
