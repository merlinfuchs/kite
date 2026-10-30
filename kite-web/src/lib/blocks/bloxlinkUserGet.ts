import { guildTargetSetting, userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const bloxlinkUserGet: BlockDefinition = {
  type: "action_bloxlink_user_get",
  id_prefix: "bloxlink",
  title: "Get Bloxlink user",
  description: "Get the Roblox user a member verified as with Bloxlink",
  icon: "link",
  category: "Roblox",
  requires: ["bloxlink", "roblox"],
  credits: 1,
  strict_settings: true,
  fields: [userTargetSetting, guildTargetSetting],
  result: { thing: "roblox_user" },
  run: { kind: "custom" },
};
