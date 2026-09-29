import { nodeActionMessagePinDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageUnpin: BlockDefinition = {
  type: "action_message_unpin",
  title: "Unpin channel message",
  description: "Bot unpins a message in a channel",
  icon: "pin-off",
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
    operation: "deprecated_delete_pin",
    method: "DELETE",
    path: "/channels/{channel_id}/pins/{message_id}",
  },
  // Only describe the request. The settings keep their schema and inputs.
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Channel",
      description: "ID of the channel.",
    },
    {
      name: "message_target",
      in: "path",
      target: "message_id",
      type: "snowflake",
      label: "Message",
      description: "ID of the message.",
    },
  ],
};
