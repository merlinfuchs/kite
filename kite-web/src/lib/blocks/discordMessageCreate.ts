import { nodeActionMessageCreateDataSchema } from "../flow/dataSchema";
import { nodeActionMessageCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordMessageCreate: BlockDefinition = {
  type: "action_message_create",
  title: "Create channel message",
  description: "Bot sends a message to a channel",
  icon: "message-circle-plus",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageCreateDataSchema,
  inputs: [
    "channel_target",
    "message_template_id",
    "message_data",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionMessageCreateResultSchema },
  run: { kind: "custom" },
};
