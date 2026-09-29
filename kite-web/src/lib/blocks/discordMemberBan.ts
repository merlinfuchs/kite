import { nodeActionMemberBanDataSchema } from "../flow/dataSchema";
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
  // Only describe the request. The settings keep their schema and inputs.
  fields: [
    {
      name: "guild_target",
      in: "path",
      target: "guild_id",
      type: "snowflake",
      label: "Server",
      description:
        "ID of the server. Leave empty to use the server the flow runs in.",
      fallback: "guild",
    },
    {
      name: "user_target",
      in: "path",
      target: "user_id",
      type: "snowflake",
      label: "User",
      description: "ID of the user.",
    },
    {
      name: "member_ban_delete_message_duration_seconds",
      in: "body",
      target: "delete_message_seconds",
      type: "integer",
      label: "Delete Message Duration",
      description:
        "Delete the member's messages from this many seconds before the ban, up to 604800 (7 days).",
      min: 0,
      max: 604800,
    },
  ],
};
