import { nodeActionMemberUnbanDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberUnban: BlockDefinition = {
  type: "action_member_unban",
  title: "Unban member",
  description: "Unban a member from the server",
  icon: "user-round-check",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberUnbanDataSchema,
  inputs: ["guild_target", "user_target", "audit_log_reason", "custom_label"],
  run: { kind: "custom" },
};
