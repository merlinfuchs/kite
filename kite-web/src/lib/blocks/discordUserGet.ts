import { userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordUserGet: BlockDefinition = {
  type: "action_user_get",
  title: "Get user",
  description: "Get a user by ID",
  icon: "user-round-search",
  category: "Users",
  requires: ["discord"],
  credits: 1,
  fields: [userTargetSetting],
  result: { thing: "discord_user" },
  run: { kind: "custom" },
};
