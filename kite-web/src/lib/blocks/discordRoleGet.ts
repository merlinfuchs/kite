import { nodeActionRoleGetResultSchema } from "../flow/resultSchema";
import { guildTargetSetting, roleTargetSetting } from "./fields";
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
  fields: [guildTargetSetting, roleTargetSetting],
  result: { schema: nodeActionRoleGetResultSchema },
  run: { kind: "custom" },
};
