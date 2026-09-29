import { nodeActionPrivateMessageCreateDataSchema } from "../flow/dataSchema";
import { nodeActionPrivateMessageCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordPrivateMessageCreate: BlockDefinition = {
  type: "action_private_message_create",
  title: "Send direct message",
  description: "Bot sends a private message to a user if the user allows it",
  icon: "message-circle-plus",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionPrivateMessageCreateDataSchema,
  inputs: [
    "user_target",
    "message_data",
    "message_template_id",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionPrivateMessageCreateResultSchema },
  run: { kind: "custom" },
};
