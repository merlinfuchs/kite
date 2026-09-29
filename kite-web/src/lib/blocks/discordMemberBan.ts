import { nodeActionMemberBanDataSchema } from "../flow/dataSchema";
import { guildField, userField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberBan: BlockDefinition = {
  type: "action_member_ban",
  title: "Ban member",
  description: "Ban a member from the server",
  icon: "user-round-x",
  category: "Members",
  credits: 1,
  schema: nodeActionMemberBanDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "member_ban_delete_message_duration_seconds",
    "audit_log_reason",
    "custom_label",
  ],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "ban_user_from_guild",
    method: "PUT",
    path: "/guilds/{guild_id}/bans/{user_id}",
  },
  fields: [
    guildField,
    userField,
    {
      name: "member_ban_delete_message_duration_seconds",
      in: "body",
      target: "delete_message_seconds",
      type: "seconds",
      min: 0,
      max: 604800,
    },
  ],
};
