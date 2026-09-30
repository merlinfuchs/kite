import { nodeActionUserGetResultSchema } from "../flow/resultSchema";
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
  allow_unknown_settings: true,
  fields: [userTargetSetting],
  result: { schema: nodeActionUserGetResultSchema },
  run: { kind: "custom" },
};
