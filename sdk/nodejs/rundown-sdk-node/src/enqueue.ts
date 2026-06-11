import { EngineClient } from "./client.js";
import type { EnqueueOptions, EnqueueResult } from "./types.js";

export async function enqueue<T = unknown>(
  queue: string,
  payload: T,
  options: EnqueueOptions & { host?: string } = {}
): Promise<EnqueueResult | null> {
  const host = options.host ?? process.env.RUNDOWN_URL || process.env.RUNDOWN_ENGINE_URL || "http://localhost:8181";
  const client = new EngineClient(host);
  return client.enqueue(queue, payload, options);
}
