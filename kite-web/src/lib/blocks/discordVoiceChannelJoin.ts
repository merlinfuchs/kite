import { z } from "zod";
import { numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceChannelJoin: BlockDefinition = {
  type: "action_voice_channel_join",
  title: "Join voice channel",
  description: "Bot joins a voice channel",
  icon: "phone-call",
  category: "Voice",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder("ID of the voice channel to join."),
    },
    {
      name: "voice_self_mute",
      schema: z.boolean().optional().describe("Whether the bot joins muted."),
    },
    {
      name: "voice_self_deaf",
      schema: z
        .boolean()
        .optional()
        .describe("Whether the bot joins deafened."),
    },
  ],
  run: { kind: "custom" },
};
