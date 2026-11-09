import { emojiDataSchema } from "../flow/dataSchema";
import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageReactionCreate: BlockDefinition = {
  type: "action_message_reaction_create",
  id_prefix: "reaction",
  title: "Create message reaction",
  description: "Bot adds a reaction to a message",
  icon: "smile-plus",
  category: "Messages",
  permissions: ["ADD_REACTIONS", "READ_MESSAGE_HISTORY"],
  run: {
    kind: "request",
    integration: "discord",
    operation: "add_my_message_reaction",
  },
  fields: [
    channelTargetField,
    messageTargetField,
    {
      name: "emoji_data",
      in: "path",
      target: "emoji_name",
      type: "emoji",
      schema: emojiDataSchema.describe("The emoji to react with."),
    },
  ],
};
