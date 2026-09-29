import { nodeActionMemberRoleRemoveDataSchema } from "../flow/dataSchema";
import { guildField, userField, roleField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberRoleRemove: BlockDefinition = {
  type: "action_member_role_remove",
  title: "Remove role from member",
  description: "Remove a role from a member",
  icon: "bookmark-minus",
  category: "Roles",
  credits: 1,
  schema: nodeActionMemberRoleRemoveDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "role_target",
    "audit_log_reason",
    "custom_label",
  ],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_guild_member_role",
    method: "DELETE",
    path: "/guilds/{guild_id}/members/{user_id}/roles/{role_id}",
  },
  fields: [guildField, userField, roleField],
};
