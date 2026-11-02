import { z } from "zod";
import { guildTargetSetting, userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberPresenceGet: BlockDefinition = {
  type: "action_member_presence_get",
  id_prefix: "presence",
  title: "Get member status",
  description:
    "Get the status and custom status of a member. Needs the Presence Intent",
  icon: "circle-dot",
  category: "Members",
  requires: ["discord"],
  credits: 0,
  strict_settings: true,
  fields: [guildTargetSetting, userTargetSetting],
  result: {
    schema: z
      .object({
        status: z
          .string()
          .describe(
            "online, idle, dnd or offline. Invisible members are offline"
          ),
        custom_status: z
          .string()
          .describe("Text of the custom status, empty without one"),
        custom_status_emoji: z
          .string()
          .describe(
            "Emoji of the custom status, a unicode emoji or a custom emoji mention"
          ),
        activity: z
          .string()
          .describe(
            "Name of what the member is doing, like the game they play, empty if nothing"
          ),
      })
      .describe("The status of the member"),
  },
  run: { kind: "custom" },
};
