import { nodeActionMemberUnbanDataSchema } from "../flow/dataSchema";
import { guildField, userField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberUnban: BlockDefinition = {
  type: "action_member_unban",
  title: "Unban member",
  description: "Unban a member from the server",
  icon: "user-round-check",
  category: "Members",
  credits: 1,
  schema: nodeActionMemberUnbanDataSchema,
  inputs: ["guild_target", "user_target", "audit_log_reason", "custom_label"],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "unban_user_from_guild",
    method: "DELETE",
    path: "/guilds/{guild_id}/bans/{user_id}",
  },
  fields: [guildField, userField],
};
