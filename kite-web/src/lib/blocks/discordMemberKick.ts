import { nodeActionMemberKickDataSchema } from "../flow/dataSchema";
import { guildField, userField } from "./fields";
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
  fields: [guildField, userField],
};
