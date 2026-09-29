import { nodeActionUserGetDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordUserGet: BlockDefinition = {
  type: "action_user_get",
  title: "Get user",
  description: "Get a user by ID",
  icon: "user-round-search",
  category: "Users",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionUserGetDataSchema,
  inputs: ["user_target", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
