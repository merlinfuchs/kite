import { nodeActionMessageReactionCreateDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageReactionCreate: BlockDefinition = {
  type: "action_message_reaction_create",
  title: "Create message reaction",
  description: "Bot adds a reaction to a message",
  icon: "smile-plus",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageReactionCreateDataSchema,
  inputs: ["channel_target", "message_target", "emoji_data", "custom_label"],
  run: { kind: "custom" },
};
