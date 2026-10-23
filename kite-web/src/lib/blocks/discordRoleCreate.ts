import { guildField, roleDataFields } from "./fields";
import { BlockDefinition } from "./types";

export const discordRoleCreate: BlockDefinition = {
  type: "action_role_create",
  id_prefix: "role",
  title: "Create role",
  description: "Create a new role in the server",
  icon: "shield-plus",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "create_guild_role",
    method: "POST",
    path: "/guilds/{guild_id}/roles",
  },
  fields: [guildField, ...roleDataFields(false)],
  result: { thing: "discord_role" },
};
