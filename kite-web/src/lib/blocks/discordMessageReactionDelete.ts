import { emojiDataSchema } from "../flow/dataSchema";
import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageReactionDelete: BlockDefinition = {
  type: "action_message_reaction_delete",
  title: "Delete message reaction",
  description: "Bot deletes a reaction from a message",
  icon: "frown",
  category: "Messages",
  credits: 1,
  allow_unknown_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_my_message_reaction",
    method: "DELETE",
    path: "/channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me",
  },
  fields: [
    channelTargetField,
    messageTargetField,
    {
      name: "emoji_data",
      in: "path",
      target: "emoji_name",
      type: "emoji",
      schema: emojiDataSchema.describe("The emoji to remove the reaction of."),
    },
  ],
};
