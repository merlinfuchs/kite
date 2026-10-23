import { guildField, roleTargetField, roleDataFields } from "./fields";
import { BlockDefinition } from "./types";

export const discordRoleEdit: BlockDefinition = {
  type: "action_role_edit",
  id_prefix: "role",
  title: "Edit role",
  description: "Edit a role in the server",
  icon: "shield-ellipsis",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "update_guild_role",
    method: "PATCH",
    path: "/guilds/{guild_id}/roles/{role_id}",
  },
  fields: [
    guildField,
    { ...roleTargetField, required: true },
    ...roleDataFields(true),
  ],
  result: { thing: "discord_role" },
};
