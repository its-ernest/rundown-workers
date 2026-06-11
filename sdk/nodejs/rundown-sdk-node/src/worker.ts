import { EngineClient } from "./client.js";
import type {
  Job,
  Handler,
  QueueConfig,
  QueueRegistration,
  WorkerConfig,
} from "./types.js";

const DEFAULT_POLL_INTERVAL = 2000;
const DEFAULT_CONCURRENCY = 1;
const DEFAULT_TIMEOUT = 300_000;

interface QueueSlot {
  registration: QueueRegistration;
  activeJobs: number;
}

export class Worker {
  private client: EngineClient;
  private config: Required<Omit<WorkerConfig, "onError">> &
    Pick<WorkerConfig, "onError">;
  private queues: Map<string, QueueSlot> = new Map();
  private running = false;
  private activeJobPromises: Set<Promise<void>> = new Set();

  constructor(config: WorkerConfig) {
    this.config = {
      host: config.host,
      pollInterval: config.pollInterval ?? DEFAULT_POLL_INTERVAL,
      concurrency: config.concurrency ?? DEFAULT_CONCURRENCY,
      onError: config.onError,
    };
    this.client = new EngineClient(config.host);
  }

  register<T>(queueConfig: QueueConfig, handler: Handler<T>): this {
    this.queues.set(queueConfig.name, {
      registration: { config: queueConfig, handler: handler as Handler },
      activeJobs: 0,
    });
    return this;
  }

  async start(): Promise<void> {
    if (this.running) return;
    this.running = true;
    this.setupGracefulShutdown();
    await Promise.all([...this.queues.keys()].map((q) => this.pollQueue(q)));
  }

  async stop(): Promise<void> {
    this.running = false;
    await Promise.all(this.activeJobPromises);
  }

  private async pollQueue(queueName: string): Promise<void> {
    while (this.running) {
      const slot = this.queues.get(queueName);
      if (!slot) break;

      const maxConcurrency =
        slot.registration.config.concurrency ?? this.config.concurrency;

      if (slot.activeJobs < maxConcurrency) {
        const job = await this.client.poll(queueName);
        if (job) {
          const p = this.executeJob(job, slot);
          this.activeJobPromises.add(p);
          p.finally(() => this.activeJobPromises.delete(p));
        } else {
          await this.sleep(this.config.pollInterval);
        }
      } else {
        await this.sleep(this.config.pollInterval);
      }
    }
  }

  private async executeJob(job: Job, slot: QueueSlot): Promise<void> {
    slot.activeJobs++;
    const timeoutMs = job.timeout ? job.timeout * 1000 : DEFAULT_TIMEOUT;

    try {
      await this.withTimeout(
        Promise.resolve(slot.registration.handler(job.payload)),
        timeoutMs,
        job.id
      );
      await this.client.complete(job.id);
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      this.config.onError?.(error, job.id, job.queue);
      console.error(
        `[rundown] job ${job.id} on "${job.queue}" failed:`,
        error.message
      );
      await this.client.fail(job.id);
    } finally {
      slot.activeJobs--;
    }
  }

  private withTimeout<T>(
    promise: Promise<T>,
    ms: number,
    jobId: string
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error(`job ${jobId} timed out after ${ms}ms`));
      }, ms);

      promise.then(
        (val) => {
          clearTimeout(timer);
          resolve(val);
        },
        (err) => {
          clearTimeout(timer);
          reject(err);
        }
      );
    });
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  private setupGracefulShutdown(): void {
    const shutdown = async () => {
      console.log("[rundown] shutting down, waiting for active jobs...");
      await this.stop();
      process.exit(0);
    };

    process.once("SIGINT", shutdown);
    process.once("SIGTERM", shutdown);
  }
}
