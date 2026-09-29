import { nodeActionVoiceChannelLeaveDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordVoiceChannelLeave: BlockDefinition = {
  type: "action_voice_channel_leave",
  title: "Leave voice channel",
  description: "Bot leaves its voice channel in a server",
  icon: "phone-off",
  category: "Voice",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionVoiceChannelLeaveDataSchema,
  inputs: ["guild_target", "custom_label"],
  run: { kind: "custom" },
};
