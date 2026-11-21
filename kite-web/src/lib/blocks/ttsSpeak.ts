import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

// Match ttsMaxLength, ttsMinSpeed, ttsMaxSpeed and ttsVoices in
// kite-service/pkg/flow/data.go.
export const ttsMaxLength = 1000;
export const ttsMinSpeed = 0.25;
export const ttsMaxSpeed = 4;

export const ttsVoices = [
  { value: "alloy", label: "Alloy" },
  { value: "ash", label: "Ash" },
  { value: "coral", label: "Coral" },
  { value: "echo", label: "Echo" },
  { value: "fable", label: "Fable" },
  { value: "nova", label: "Nova" },
  { value: "onyx", label: "Onyx" },
  { value: "sage", label: "Sage" },
  { value: "shimmer", label: "Shimmer" },
] as const;

export const ttsDefaultVoice = "alloy";

// The speeds of the editor's select. The service accepts any speed in range,
// and normal speed is stored as no speed.
export const ttsNormalSpeed = "1";
export const ttsSpeeds = [
  { value: "0.5", label: "Very slow (0.5x)" },
  { value: "0.75", label: "Slow (0.75x)" },
  { value: ttsNormalSpeed, label: "Normal (1x)" },
  { value: "1.25", label: "Fast (1.25x)" },
  { value: "1.5", label: "Faster (1.5x)" },
  { value: "2", label: "Very fast (2x)" },
];

const n = (v: number) => v.toLocaleString("en-US");

// For the descriptions of the text in the block and its editor.
export const ttsLimits = `Up to ${n(
  ttsMaxLength
)} characters with placeholders filled in.`;

export const ttsSpeak: BlockDefinition = {
  type: "action_tts_speak",
  id_prefix: "speak",
  title: "Speak Text",
  description: "Bot speaks text in a voice channel",
  icon: "audio-lines",
  category: "Voice",
  requires: ["discord"],
  premium_feature: "voice_audio",
  credits: 2,
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
      name: "tts_data",
      schema: z
        .object({
          text: templated(
            z.string().min(1).max(ttsMaxLength),
            `What to say, like {{message.content}}. ${ttsLimits}`
          ),
          voice: z
            .enum(ttsVoices.map((v) => v.value) as [string, ...string[]])
            .optional()
            .describe(`The voice that speaks. Defaults to ${ttsDefaultVoice}.`),
          speed: z
            .number()
            .min(ttsMinSpeed)
            .max(ttsMaxSpeed)
            .optional()
            .describe(
              `How fast the text is spoken, from ${ttsMinSpeed} to ${ttsMaxSpeed}. Defaults to normal speed, 1.`
            ),
        })
        .describe("The text to speak and how it sounds."),
    },
    {
      name: "audio_leave_after",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether the bot leaves the voice channel once it has spoken. Otherwise it stays until it's told to leave."
        ),
    },
  ],
  run: { kind: "custom" },
};
