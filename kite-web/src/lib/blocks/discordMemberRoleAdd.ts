import { nodeActionMemberRoleAddDataSchema } from "../flow/dataSchema";
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
      name: "role_target",
      in: "path",
      target: "role_id",
      type: "snowflake",
      label: "Role",
      description: "ID of the role.",
    },
  ],
};
