import { nodeActionMemberBanDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberBan: BlockDefinition = {
  type: "action_member_ban",
  title: "Ban member",
  description: "Ban a member from the server",
  icon: "user-round-x",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberBanDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "member_ban_delete_message_duration_seconds",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
