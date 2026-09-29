import { roleResultSchema } from "../../flow/resultSchema";
import { BlockDefinition } from "../types";

export const createRole: BlockDefinition = {
  type: "action_role_create",
  title: "Create role",
  description: "Create a new role in the server",
  icon: "shield-plus",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "create_guild_role",
    method: "POST",
    path: "/guilds/{guild_id}/roles",
  },
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
      name: "name",
      in: "body",
      type: "string",
      label: "Name",
      description: "Name of the role. Defaults to 'new role'.",
      max_length: 100,
    },
    {
      name: "permissions",
      in: "body",
      type: "string",
      label: "Permissions",
      description:
        "Permissions of the role. Defaults to the ones of @everyone.",
      widget: "permissions",
    },
    {
      name: "color",
      in: "body",
      type: "integer",
      label: "Color",
      description:
        "Color of the role as a number, e.g. 16711680 for red. Defaults to no color.",
      min: 0,
      max: 16777215,
    },
    {
      name: "hoist",
      in: "body",
      type: "boolean",
      label: "Show Separately",
      description:
        "Whether members with the role are shown separately in the member list.",
    },
    {
      name: "mentionable",
      in: "body",
      type: "boolean",
      label: "Mentionable",
      description: "Whether everyone can mention the role.",
    },
  ],
  result: { thing: "discord_role", schema: roleResultSchema },
};
