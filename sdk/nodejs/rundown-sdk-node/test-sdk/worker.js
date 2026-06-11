// worker.js
import { queue, run } from "rundown-workers";

queue({ name: "new email check" }, async (payload) => {
  console.log("processing job:", payload);
});

await run({ host: process.env.RUNDOWN_HOST ?? "http://localhost:8181" });
