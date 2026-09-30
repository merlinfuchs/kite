import { channelTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordChannelDelete: BlockDefinition = {
  type: "action_channel_delete",
  title: "Delete channel",
  description: "Delete a channel or thread ",
  icon: "folder-x",
  category: "Channels",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_channel",
    method: "DELETE",
    path: "/channels/{channel_id}",
  },
  fields: [channelTargetField],
};
