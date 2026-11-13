import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceTTSSpeak: BlockDefinition = {
  type: "action_voice_tts_speak",
  id_prefix: "speak",
  title: "Speak text",
  description: "Bot reads a text out loud in a voice channel",
  icon: "mic",
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
        "ID of the voice channel to speak in. The bot joins it, or moves there from another channel of the server."
      ),
    },
    {
      name: "tts_text",
      schema: templated(
        z
          .string()
          .min(1, "Enter the text to speak")
          .describe(
            "What the bot says, up to 500 characters. Placeholders like {{arg('text')}} work."
          ),
        "What the bot says, up to 500 characters. Placeholders like {{arg('text')}} work."
      ),
    },
    {
      name: "tts_language",
      schema: templated(
        z
          .string()
          .optional()
          .describe(
            "The language the text is read in, as a code like en, de, fr, es or pt-BR. Defaults to English."
          ),
        "The language the text is read in, as a code like en, de, fr, es or pt-BR. Defaults to English."
      ),
    },
    {
      name: "audio_leave_after",
      input: "audio_leave_after",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether the bot leaves the voice channel once it finished speaking. Otherwise it stays until it's told to leave."
        ),
    },
  ],
  run: { kind: "custom" },
};
