import { nodeActionMessageReactionDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageReactionDelete: BlockDefinition = {
  type: "action_message_reaction_delete",
  title: "Delete message reaction",
  description: "Bot deletes a reaction from a message",
  icon: "frown",
  category: "Messages",
  credits: 1,
  schema: nodeActionMessageReactionDeleteDataSchema,
  inputs: ["channel_target", "message_target", "emoji_data", "custom_label"],
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_my_message_reaction",
    method: "DELETE",
    path: "/channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me",
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
    {
      name: "emoji_data",
      in: "path",
      target: "emoji_name",
      type: "emoji",
      label: "Emoji",
      description: "The emoji.",
    },
  ],
};
