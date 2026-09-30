import { guildTargetSchema, roleTargetSchema } from "../flow/dataSchema";
import { nodeActionRoleGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordRoleGet: BlockDefinition = {
  type: "action_role_get",
  title: "Get role",
  description: "Get a role by ID",
  icon: "bookmark",
  category: "Roles",
  requires: ["discord"],
  credits: 1,
  allow_unknown_settings: true,
  fields: [
    {
      name: "guild_target",
      type: "snowflake",
      schema: guildTargetSchema.optional(),
    },
    { name: "role_target", type: "snowflake", schema: roleTargetSchema },
  ],
  result: { schema: nodeActionRoleGetResultSchema },
  run: { kind: "custom" },
};
