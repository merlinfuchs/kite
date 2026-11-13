import { emojiDataSchema } from "../flow/dataSchema";
import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageReactionDelete: BlockDefinition = {
  type: "action_message_reaction_delete",
  id_prefix: "reaction",
  title: "Delete message reaction",
  description: "Bot deletes a reaction from a message",
  icon: "frown",
  category: "Messages",
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_my_message_reaction",
  },
  fields: [
    channelTargetField,
    messageTargetField,
    {
      name: "emoji_data",
      target: "emoji_name",
      type: "emoji",
      schema: emojiDataSchema.describe("The emoji to remove the reaction of."),
    },
  ],
};
