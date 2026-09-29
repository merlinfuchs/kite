import { nodeActionThreadCreateDataSchema } from "../flow/dataSchema";
import { nodeActionThreadCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordThreadCreate: BlockDefinition = {
  type: "action_thread_create",
  title: "Create thread",
  description: "Create a thread",
  icon: "message-circle-plus",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionThreadCreateDataSchema,
  inputs: ["thread_data", "audit_log_reason", "temporary_name", "custom_label"],
  result: { schema: nodeActionThreadCreateResultSchema },
  run: { kind: "custom" },
};
