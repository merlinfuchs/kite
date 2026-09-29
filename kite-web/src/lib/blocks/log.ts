import { nodeActionLogDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const log: BlockDefinition = {
  type: "action_log",
  title: "Log Message",
  description: "Log some text which is only visible in the application logs",
  icon: "scroll-text",
  category: "Utilities",
  credits: 1,
  schema: nodeActionLogDataSchema,
  inputs: ["log_level", "log_message", "custom_label"],
  run: { kind: "custom" },
};
