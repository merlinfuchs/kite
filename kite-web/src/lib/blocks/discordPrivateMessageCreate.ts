import { numericOrPlaceholder } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
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
  fields: [
    {
      name: "user_target",
      schema: numericOrPlaceholder("ID of the user to send the message to."),
    },
    messageDataField,
    messageTemplateField,
  ],
  refine: requireMessage,
  result: { schema: nodeActionPrivateMessageCreateResultSchema },
  run: { kind: "custom" },
};
