import { guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordVoiceChannelLeave: BlockDefinition = {
  type: "action_voice_channel_leave",
  title: "Leave voice channel",
  description: "Bot leaves its voice channel in a server",
  icon: "phone-off",
  category: "Voice",
  requires: ["discord"],
  credits: 1,
  allow_unknown_settings: true,
  fields: [guildTargetSetting],
  run: { kind: "custom" },
};
