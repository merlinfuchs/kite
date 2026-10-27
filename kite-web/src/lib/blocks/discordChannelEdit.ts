import { channelEditDataSchema } from "../flow/dataSchema";
import { channelDataSetting, channelTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordChannelEdit: BlockDefinition = {
  type: "action_channel_edit",
  id_prefix: "channel",
  title: "Edit channel",
  description: "Edit a channel or thread",
  icon: "folder-pen",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  permissions: ["Manage Channels"],
  fields: [
    channelTargetSetting,
    {
      ...channelDataSetting,
      input: "channel_edit_data",
      schema: channelEditDataSchema,
    },
  ],
  audit_log_reason: true,
  result: { thing: "discord_channel" },
  run: { kind: "custom" },
};
