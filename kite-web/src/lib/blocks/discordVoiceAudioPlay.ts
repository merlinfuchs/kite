import { z } from "zod";
import {
  decimalRegex,
  numericOrPlaceholder,
  templated,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceAudioPlay: BlockDefinition = {
  type: "action_voice_audio_play",
  id_prefix: "audio",
  title: "Play audio",
  description: "Bot plays an audio file in a voice channel",
  icon: "volume-2",
  category: "Voice",
  requires: ["discord"],
  // Downloads the file and converts it while it plays.
  credits: 5,
  permissions: ["CONNECT", "SPEAK"],
  strict_settings: true,
  fields: [
    {
      name: "channel_target",
      input: "audio_channel_target",
      schema: numericOrPlaceholder(
        "ID of the voice channel to play the audio in. The bot joins it, or moves there from another channel of the server."
      ),
    },
    {
      name: "audio_url",
      schema: templated(
        z
          .string()
          .regex(
            /^\s*(https?:\/\/|\{\{)/,
            "Must be a link starting with http:// or https://, or a placeholder"
          ),
        "Link to the audio file, like an MP3, WAV or OGG file. For a file uploaded with an attachment command argument, use its url, like {{arg('sound').url}}. Up to 25 MB, of which the first 15 minutes play."
      ),
    },
    {
      name: "audio_volume",
      schema: numericOrPlaceholder(
        "Volume in percent, from 1 to 200. Defaults to 100, the volume of the file.",
        decimalRegex
      ).optional(),
    },
    {
      name: "audio_leave_after",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether the bot leaves the voice channel once the audio ended. Otherwise it stays until it's told to leave."
        ),
    },
  ],
  run: { kind: "custom" },
};
