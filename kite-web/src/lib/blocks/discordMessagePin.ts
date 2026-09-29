import { nodeActionMessagePinDataSchema } from "../flow/dataSchema";
import { channelField, messageField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessagePin: BlockDefinition = {
  type: "action_message_pin",
  title: "Pin channel message",
  description: "Bot pins a message in a channel",
  icon: "pin",
  category: "Messages",
  credits: 1,
  schema: nodeActionMessagePinDataSchema,
  inputs: [
    "channel_target",
    "message_target",
    "audit_log_reason",
    "custom_label",
  ],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "deprecated_create_pin",
    method: "PUT",
    path: "/channels/{channel_id}/pins/{message_id}",
  },
  fields: [channelField, messageField],
};
