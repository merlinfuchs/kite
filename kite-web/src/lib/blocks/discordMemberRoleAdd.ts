import { nodeActionMemberRoleAddDataSchema } from "../flow/dataSchema";
import { guildField, userField, roleField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberRoleAdd: BlockDefinition = {
  type: "action_member_role_add",
  title: "Add role to member",
  description: "Add a role to a member",
  icon: "bookmark-plus",
  category: "Roles",
  credits: 1,
  schema: nodeActionMemberRoleAddDataSchema,
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
    operation: "add_guild_member_role",
    method: "PUT",
    path: "/guilds/{guild_id}/members/{user_id}/roles/{role_id}",
  },
  fields: [guildField, userField, roleField],
};
