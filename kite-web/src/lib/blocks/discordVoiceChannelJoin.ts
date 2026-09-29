import { nodeActionVoiceChannelJoinDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceChannelJoin: BlockDefinition = {
  type: "action_voice_channel_join",
  title: "Join voice channel",
  description: "Bot joins a voice channel",
  icon: "phone-call",
  category: "Voice",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionVoiceChannelJoinDataSchema,
  inputs: [
    "channel_target",
    "voice_self_mute",
    "voice_self_deaf",
    "custom_label",
  ],
  run: { kind: "custom" },
};
