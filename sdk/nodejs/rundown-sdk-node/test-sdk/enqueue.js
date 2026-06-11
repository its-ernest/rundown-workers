import { enqueue } from "rundown-workers";

const result = await enqueue("new email check", { to: "user@123example.com" });
console.log("enqueued:", result);
