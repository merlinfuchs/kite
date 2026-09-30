import { userTargetSchema } from "../flow/dataSchema";
import { nodeActionUserGetResultSchema } from "../flow/resultSchema";
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
  fields: [
    { name: "user_target", type: "snowflake", schema: userTargetSchema },
  ],
  result: { schema: nodeActionUserGetResultSchema },
  run: { kind: "custom" },
};
