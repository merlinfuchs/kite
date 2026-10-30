import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockField } from "./types";

// Settings of the Roblox group blocks.

export const robloxGroupIdSetting: BlockField = {
  name: "roblox_group_id",
  schema: numericOrPlaceholder(
    "ID of the Roblox group, the number in the group's link."
  ),
};

export const robloxGroupUserSettings: BlockField[] = [
  {
    name: "roblox_user_target",
    schema: templated(
      z.string(),
      "Roblox ID or username of the user, depending on roblox_lookup_mode. A Roblox user result works too. Discord users and IDs don't work, use action_bloxlink_user_get to find the Roblox user of a member."
    ),
  },
  {
    name: "roblox_lookup_mode",
    schema: z
      .enum(["id", "username"])
      .optional()
      .describe("Whether roblox_user_target is an ID or a username."),
  },
];

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
