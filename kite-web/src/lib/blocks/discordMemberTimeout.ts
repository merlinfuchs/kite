import { nodeActionMemberTimeoutDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberTimeout: BlockDefinition = {
  type: "action_member_timeout",
  title: "Timeout member",
  description: "Timeout a member in the server",
  icon: "message-circle-off",
  category: "Members",
  credits: 1,
  schema: nodeActionMemberTimeoutDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "member_timeout_duration_seconds",
    "audit_log_reason",
    "custom_label",
  ],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "update_guild_member",
    method: "PATCH",
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
    {
      name: "member_timeout_duration_seconds",
      in: "body",
      target: "communication_disabled_until",
      type: "seconds_until",
      label: "Timeout Duration",
      description:
        "How many seconds the member is timed out for, up to 2419200 (28 days).",
      required: true,
      min: 0,
      max: 2419200,
    },
  ],
};
