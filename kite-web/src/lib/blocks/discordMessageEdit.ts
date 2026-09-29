import { nodeActionMessageEditDataSchema } from "../flow/dataSchema";
import { nodeActionMessageEditResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordMessageEdit: BlockDefinition = {
  type: "action_message_edit",
  title: "Edit channel message",
  description: "Bot edits an existing message in a channel",
  icon: "pen",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageEditDataSchema,
  inputs: [
    "channel_target",
    "message_target",
    "message_template_id",
    "message_data",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionMessageEditResultSchema },
  run: { kind: "custom" },
};
