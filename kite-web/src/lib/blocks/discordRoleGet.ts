import { guildTargetSetting, roleTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordRoleGet: BlockDefinition = {
  type: "action_role_get",
  id_prefix: "role",
  title: "Get role",
  description: "Get a role by ID",
  icon: "bookmark",
  category: "Roles",
  requires: ["discord"],
  credits: 0,
  fields: [guildTargetSetting, roleTargetSetting],
  result: { thing: "discord_role" },
  run: { kind: "custom" },
};
