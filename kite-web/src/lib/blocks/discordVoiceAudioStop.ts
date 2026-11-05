import { guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordVoiceAudioStop: BlockDefinition = {
  type: "action_voice_audio_stop",
  id_prefix: "audio",
  title: "Stop audio",
  description: "Bot stops playing audio in a server and stays in the channel",
  icon: "volume-x",
  category: "Voice",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [guildTargetSetting],
  run: { kind: "custom" },
};
