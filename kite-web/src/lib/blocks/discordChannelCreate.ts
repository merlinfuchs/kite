import { channelDataSetting, guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordChannelCreate: BlockDefinition = {
  type: "action_channel_create",
  title: "Create channel",
  description: "Create a channel",
  icon: "folder-plus",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [guildTargetSetting, channelDataSetting],
  audit_log_reason: true,
  result: { thing: "discord_channel" },
  run: { kind: "custom" },
};
