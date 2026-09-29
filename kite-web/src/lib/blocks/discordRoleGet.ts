import { nodeActionRoleGetDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionRoleGetDataSchema,
  inputs: ["guild_target", "role_target", "temporary_name", "custom_label"],
  result: { schema: nodeActionRoleGetResultSchema },
  run: { kind: "custom" },
};
