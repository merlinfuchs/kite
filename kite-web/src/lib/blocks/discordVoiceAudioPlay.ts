import { z } from "zod";
import { numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceAudioPlay: BlockDefinition = {
  type: "action_voice_audio_play",
  id_prefix: "audio",
  title: "Play audio",
  description: "Bot plays an uploaded sound in a voice channel",
  icon: "volume-2",
  category: "Voice",
  requires: ["discord"],
  premium_feature: "voice_audio",
  credits: 1,
  permissions: ["CONNECT", "SPEAK"],
  strict_settings: true,
  fields: [
    {
      name: "channel_target",
      input: "audio_channel_target",
      schema: numericOrPlaceholder(
        "ID of the voice channel to play the sound in. The bot joins it, or moves there from another channel of the server."
      ),
    },
    {
      name: "audio_asset_id",
      schema: z
        .string()
        .min(1, "Upload a sound to play")
        .describe(
          "ID of the sound to play. Sounds are uploaded in the block's settings in the editor, up to 2 minutes long and 8 MB."
        ),
    },
    {
      name: "audio_leave_after",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether the bot leaves the voice channel once the sound ended. Otherwise it stays until it's told to leave, or until it played nothing for 15 minutes."
        ),
    },
  ],
  run: { kind: "custom" },
};
