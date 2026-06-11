import type { Job, EnqueueOptions, EnqueueResult } from "./types.js";

export class EngineClient {
  private host: string;
  private headers: Record<string, string>;

  constructor(host: string, headers: Record<string, string> = {}) {
    this.host = host.replace(/\/$/, "");
    this.headers = {
      "Content-Type": "application/json",
      ...headers,
    };
  }

  private async post<T>(path: string, body: unknown): Promise<T | null> {
    try {
      const res = await fetch(`${this.host}${path}`, {
        method: "POST",
        headers: this.headers,
        body: JSON.stringify(body),
      });

      if (!res.ok) return null;

      const text = await res.text();
      if (!text) return null;

      return JSON.parse(text) as T;
    } catch {
      return null;
    }
  }

  async enqueue(
    queue: string,
    payload: unknown,
    options: EnqueueOptions = {}
  ): Promise<EnqueueResult | null> {
    return this.post<EnqueueResult>("/enqueue", {
      queue,

      payload: typeof payload === "string" ? payload : JSON.stringify(payload),
      ...(options.tag !== undefined && { tag: options.tag }),
      ...(options.timeout !== undefined && { timeout: options.timeout }),
      ...(options.max_retries !== undefined && {
        max_retries: options.max_retries,
      }),
    });
  }

  async poll(queue: string): Promise<Job | null> {
    const job = await this.post<Job>("/poll", { queue });
    if (!job) return null;
    if (typeof job.payload === "string") {
      try {
        job.payload = JSON.parse(job.payload);
      } catch {}
    }
    return job;
  }

  async complete(id: string): Promise<void> {
    await this.post("/complete", { id });
  }

  async fail(id: string): Promise<void> {
    await this.post("/fail", { id });
  }
}
