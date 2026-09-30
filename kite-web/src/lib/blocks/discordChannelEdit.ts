import { nodeActionChannelEditResultSchema } from "../flow/resultSchema";
import { channelDataSetting, channelTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordChannelEdit: BlockDefinition = {
  type: "action_channel_edit",
  title: "Edit channel",
  description: "Edit a channel or thread",
  icon: "folder-pen",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [channelTargetSetting, channelDataSetting],
  audit_log_reason: true,
  result: { schema: nodeActionChannelEditResultSchema },
  run: { kind: "custom" },
};
