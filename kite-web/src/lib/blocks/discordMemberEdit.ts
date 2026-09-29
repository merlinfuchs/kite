import { nodeActionMemberEditDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberEdit: BlockDefinition = {
  type: "action_member_edit",
  title: "Edit member nickname",
  description: "Edit a member in the server",
  icon: "user-round-pen",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberEditDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "member_nick",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
