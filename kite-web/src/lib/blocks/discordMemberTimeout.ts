import { nodeActionMemberTimeoutDataSchema } from "../flow/dataSchema";
import { guildField, userField } from "./fields";
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
    partial: true,
  },
  fields: [
    guildField,
    userField,
    {
      name: "member_timeout_duration_seconds",
      in: "body",
      target: "communication_disabled_until",
      type: "seconds_until",
      required: true,
      min: 0,
      max: 2419200,
    },
  ],
};
