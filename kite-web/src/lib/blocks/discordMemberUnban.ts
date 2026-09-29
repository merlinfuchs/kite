import { nodeActionMemberUnbanDataSchema } from "../flow/dataSchema";
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
  ],
};
