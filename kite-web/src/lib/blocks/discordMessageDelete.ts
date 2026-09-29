import { nodeActionMessageDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageDelete: BlockDefinition = {
  type: "action_message_delete",
  title: "Delete channel message",
  description: "Bot deletes an existing message in a channel",
  icon: "message-circle-x",
  category: "Messages",
  credits: 1,
  schema: nodeActionMessageDeleteDataSchema,
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
    operation: "delete_message",
    method: "DELETE",
    path: "/channels/{channel_id}/messages/{message_id}",
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
