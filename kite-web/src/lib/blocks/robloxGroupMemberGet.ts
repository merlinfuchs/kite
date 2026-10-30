import {
  robloxGroupIdSetting,
  robloxGroupMemberSchema,
  robloxGroupUserSettings,
} from "./robloxGroupFields";
import { BlockDefinition } from "./types";

export const robloxGroupMemberGet: BlockDefinition = {
  type: "action_roblox_group_member_get",
  id_prefix: "member",
  title: "Get Roblox group member",
  description: "Get the roles of a user in your Roblox group",
  icon: "users",
  category: "Roblox",
  requires: ["roblox_open_cloud"],
  credits: 1,
  strict_settings: true,
  fields: [robloxGroupIdSetting, ...robloxGroupUserSettings],
  result: { schema: robloxGroupMemberSchema },
  run: { kind: "custom" },
};
