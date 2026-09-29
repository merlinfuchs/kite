import { nodeActionMessageReactionDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageReactionDelete: BlockDefinition = {
  type: "action_message_reaction_delete",
  title: "Delete message reaction",
  description: "Bot deletes a reaction from a message",
  icon: "frown",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageReactionDeleteDataSchema,
  inputs: ["channel_target", "message_target", "emoji_data", "custom_label"],
  run: { kind: "custom" },
};
