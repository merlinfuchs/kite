import { nodeActionMessageReactionCreateDataSchema } from "../flow/dataSchema";
import { channelField, messageField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageReactionCreate: BlockDefinition = {
  type: "action_message_reaction_create",
  title: "Create message reaction",
  description: "Bot adds a reaction to a message",
  icon: "smile-plus",
  category: "Messages",
  credits: 1,
  schema: nodeActionMessageReactionCreateDataSchema,
  inputs: ["channel_target", "message_target", "emoji_data", "custom_label"],
  run: {
    kind: "request",
    integration: "discord",
    operation: "add_my_message_reaction",
    method: "PUT",
    path: "/channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me",
  },
  fields: [
    channelField,
    messageField,
    {
      name: "emoji_data",
      in: "path",
      target: "emoji_name",
      type: "emoji",
    },
  ],
};
