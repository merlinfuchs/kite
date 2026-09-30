import { guildTargetField, userTargetField, roleTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberRoleRemove: BlockDefinition = {
  type: "action_member_role_remove",
  title: "Remove role from member",
  description: "Remove a role from a member",
  icon: "bookmark-minus",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  allow_unknown_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_guild_member_role",
    method: "DELETE",
    path: "/guilds/{guild_id}/members/{user_id}/roles/{role_id}",
  },
  fields: [guildTargetField, userTargetField, roleTargetField],
};
