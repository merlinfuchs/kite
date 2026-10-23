import { guildField, roleTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordRoleDelete: BlockDefinition = {
  type: "action_role_delete",
  id_prefix: "role",
  title: "Delete role",
  description: "Delete a role from the server",
  icon: "shield-x",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_guild_role",
    method: "DELETE",
    path: "/guilds/{guild_id}/roles/{role_id}",
  },
  fields: [guildField, { ...roleTargetField, required: true }],
};
