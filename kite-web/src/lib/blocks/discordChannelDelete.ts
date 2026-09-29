import { nodeActionChannelDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordChannelDelete: BlockDefinition = {
  type: "action_channel_delete",
  title: "Delete channel",
  description: "Delete a channel or thread ",
  icon: "folder-x",
  category: "Channels",
  credits: 1,
  schema: nodeActionChannelDeleteDataSchema,
  inputs: ["channel_target", "audit_log_reason", "custom_label"],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_channel",
    method: "DELETE",
    path: "/channels/{channel_id}",
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
  ],
};
