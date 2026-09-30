import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageDelete: BlockDefinition = {
  type: "action_message_delete",
  title: "Delete channel message",
  description: "Bot deletes an existing message in a channel",
  icon: "message-circle-x",
  category: "Messages",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_message",
    method: "DELETE",
    path: "/channels/{channel_id}/messages/{message_id}",
  },
  fields: [channelTargetField, messageTargetField],
};
