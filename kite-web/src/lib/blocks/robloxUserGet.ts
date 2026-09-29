import { nodeActionRobloxUserGetDataSchema } from "../flow/dataSchema";
import { nodeActionRobloxUserGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const robloxUserGet: BlockDefinition = {
  type: "action_roblox_user_get",
  title: "Get Roblox User",
  description: "Get a Roblox user by ID or username",
  icon: "gamepad",
  category: "Roblox",
  requires: ["roblox"],
  credits: 1,
  schema: nodeActionRobloxUserGetDataSchema,
  inputs: [
    "roblox_user_target",
    "roblox_lookup_mode",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionRobloxUserGetResultSchema },
  run: { kind: "custom" },
};
