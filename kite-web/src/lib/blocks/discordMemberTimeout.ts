import { numericOrPlaceholder } from "../flow/dataSchema";
import { guildTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberTimeout: BlockDefinition = {
  type: "action_member_timeout",
  title: "Timeout member",
  description: "Timeout a member in the server",
  icon: "message-circle-off",
  category: "Members",
  credits: 1,
  audit_log_reason: true,
  allow_unknown_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "update_guild_member",
    method: "PATCH",
    path: "/guilds/{guild_id}/members/{user_id}",
    partial: true,
  },
  fields: [
    guildTargetField,
    userTargetField,
    {
      name: "member_timeout_duration_seconds",
      in: "body",
      target: "communication_disabled_until",
      type: "seconds_until",
      required: true,
      min: 0,
      max: 2419200,
      schema: numericOrPlaceholder(
        "How many seconds the member is timed out for."
      ),
    },
  ],
};
