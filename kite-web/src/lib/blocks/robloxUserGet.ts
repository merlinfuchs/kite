import { robloxUserSettings } from "./fields";
import { BlockDefinition } from "./types";

export const robloxUserGet: BlockDefinition = {
  type: "action_roblox_user_get",
  id_prefix: "roblox",
  title: "Get Roblox User",
  description: "Get a Roblox user by ID or username",
  icon: "gamepad",
  category: "Roblox",
  requires: ["roblox"],
  credits: 1,
  fields: robloxUserSettings,
  result: { thing: "roblox_user" },
  run: { kind: "custom" },
};
