import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { robloxGroupIdSetting, robloxUserSettings, values } from "./fields";
import { robloxGroupMemberSchema } from "./robloxGroupMemberGet";
import { BlockDefinition } from "./types";

export const robloxGroupRoleModes = [
  { value: "set", label: "Set as only role (rank)" },
  { value: "add", label: "Add role" },
  { value: "remove", label: "Remove role" },
] as const;

export const robloxGroupMemberRoleEdit: BlockDefinition = {
  type: "action_roblox_group_member_role_edit",
  id_prefix: "rank",
  title: "Change Roblox group role",
  description: "Rank a member of your Roblox group, or give or take a role",
  icon: "chevrons-up",
  category: "Roblox",
  requires: ["roblox_open_cloud", "roblox"],
  credits: 1,
  strict_settings: true,
  fields: [
    robloxGroupIdSetting,
    ...robloxUserSettings,
    {
      name: "roblox_group_role",
      schema: templated(
        z.string().min(1),
        "Name, rank (1 to 255) or ID of the role. Names are matched first, ignoring case."
      ),
    },
    {
      name: "roblox_group_role_mode",
      schema: z
        .enum(values(robloxGroupRoleModes))
        .optional()
        .describe(
          "set makes the role the member's only role, like ranking them. add and remove give or take it and keep their other roles. Defaults to set."
        ),
    },
  ],
  result: {
    schema: robloxGroupMemberSchema.describe("The member's roles afterwards"),
  },
  run: { kind: "custom" },
};
