import { nodeActionMemberKickDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberKick: BlockDefinition = {
  type: "action_member_kick",
  title: "Kick member",
  description: "Kick a member from the server",
  icon: "user-round-minus",
  category: "Members",
  credits: 1,
  schema: nodeActionMemberKickDataSchema,
  inputs: ["guild_target", "user_target", "audit_log_reason", "custom_label"],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_guild_member",
    method: "DELETE",
    path: "/guilds/{guild_id}/members/{user_id}",
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
