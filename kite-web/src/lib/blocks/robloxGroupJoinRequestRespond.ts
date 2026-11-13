import { z } from "zod";
import { robloxGroupIdSetting, robloxUserSettings, values } from "./fields";
import { BlockDefinition } from "./types";

export const robloxJoinRequestActions = [
  { value: "accept", label: "Accept" },
  { value: "decline", label: "Decline" },
] as const;

export const robloxGroupJoinRequestRespond: BlockDefinition = {
  type: "action_roblox_group_join_request_respond",
  id_prefix: "join",
  title: "Answer Roblox join request",
  description: "Accept or decline a user's request to join your Roblox group",
  icon: "user-check",
  category: "Roblox",
  requires: ["roblox_open_cloud", "roblox"],
  credits: 1,
  strict_settings: true,
  fields: [
    robloxGroupIdSetting,
    ...robloxUserSettings,
    {
      name: "roblox_join_request_action",
      schema: z
        .enum(values(robloxJoinRequestActions))
        .optional()
        .describe("Whether to accept or decline. Defaults to accept."),
    },
  ],
  result: {
    schema: z
      .boolean()
      .describe(
        "Whether the user had a join request, false if they didn't ask to join or are already in the group"
      ),
  },
  run: { kind: "custom" },
};
