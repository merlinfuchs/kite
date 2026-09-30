import { guildTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberUnban: BlockDefinition = {
  type: "action_member_unban",
  title: "Unban member",
  description: "Unban a member from the server",
  icon: "user-round-check",
  category: "Members",
  credits: 1,
  audit_log_reason: true,
  allow_unknown_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "unban_user_from_guild",
    method: "DELETE",
    path: "/guilds/{guild_id}/bans/{user_id}",
  },
  fields: [guildTargetField, userTargetField],
};
