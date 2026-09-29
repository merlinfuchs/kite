import { nodeActionMemberTimeoutDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberTimeout: BlockDefinition = {
  type: "action_member_timeout",
  title: "Timeout member",
  description: "Timeout a member in the server",
  icon: "message-circle-off",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberTimeoutDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "member_timeout_duration_seconds",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
