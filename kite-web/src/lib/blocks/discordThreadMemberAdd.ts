import { nodeActionThreadMemberAddDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordThreadMemberAdd: BlockDefinition = {
  type: "action_thread_member_add",
  title: "Add member to thread",
  description: "Add a member to a thread",
  icon: "user-plus",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionThreadMemberAddDataSchema,
  inputs: ["channel_target", "user_target", "audit_log_reason", "custom_label"],
  run: { kind: "custom" },
};
