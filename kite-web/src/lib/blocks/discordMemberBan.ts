import { numericOrPlaceholder } from "../flow/dataSchema";
import { guildTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberBan: BlockDefinition = {
  type: "action_member_ban",
  title: "Ban member",
  description: "Ban a member from the server",
  icon: "user-round-x",
  category: "Members",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "ban_user_from_guild",
    method: "PUT",
    path: "/guilds/{guild_id}/bans/{user_id}",
  },
  fields: [
    guildTargetField,
    userTargetField,
    {
      name: "member_ban_delete_message_duration_seconds",
      in: "body",
      target: "delete_message_seconds",
      type: "seconds",
      min: 0,
      max: 604800,
      schema: numericOrPlaceholder(
        "Delete the member's messages from this many seconds before the ban."
      ).optional(),
    },
  ],
};
