import { z } from "zod";
import { robloxGroupIdSetting, robloxUserSettings } from "./fields";
import { BlockDefinition } from "./types";

// A member's roles, highest first, like the group member blocks return them.
export const robloxGroupMemberSchema = z
  .object({
    user_id: z.string().describe("Roblox ID of the user"),
    role: z.string().describe("Name of the member's highest role"),
    rank: z.number().describe("Rank of the member's highest role, 1 to 255"),
    roles: z
      .array(
        z.object({
          id: z.string(),
          name: z.string(),
          rank: z.number().describe("1 to 255"),
        })
      )
      .describe("All roles of the member, highest first"),
  })
  .describe("The member, or null if the user isn't in the group");

export const robloxGroupMemberGet: BlockDefinition = {
  type: "action_roblox_group_member_get",
  id_prefix: "member",
  title: "Get Roblox group member",
  description: "Get the roles of a user in your Roblox group",
  icon: "users",
  category: "Roblox",
  requires: ["roblox_open_cloud", "roblox"],
  credits: 1,
  strict_settings: true,
  fields: [robloxGroupIdSetting, ...robloxUserSettings],
  result: { schema: robloxGroupMemberSchema },
  run: { kind: "custom" },
};
