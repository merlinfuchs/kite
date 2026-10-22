import { channelTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordChannelGet: BlockDefinition = {
  type: "action_channel_get",
  title: "Get channel",
  description: "Get a channel by ID",
  icon: "folder-search",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [channelTargetSetting],
  result: { thing: "discord_channel" },
  run: { kind: "custom" },
};
