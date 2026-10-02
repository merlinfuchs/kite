import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageUnpin: BlockDefinition = {
  type: "action_message_unpin",
  title: "Unpin channel message",
  description: "Bot unpins a message in a channel",
  icon: "pin-off",
  category: "Messages",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "deprecated_delete_pin",
    method: "DELETE",
    path: "/channels/{channel_id}/pins/{message_id}",
  },
  fields: [channelTargetField, messageTargetField],
};
