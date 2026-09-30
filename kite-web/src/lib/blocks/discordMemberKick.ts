import { guildTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberKick: BlockDefinition = {
  type: "action_member_kick",
  title: "Kick member",
  description: "Kick a member from the server",
  icon: "user-round-minus",
  category: "Members",
  credits: 1,
  audit_log_reason: true,
  allow_unknown_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_guild_member",
    method: "DELETE",
    path: "/guilds/{guild_id}/members/{user_id}",
  },
  fields: [guildTargetField, userTargetField],
};
