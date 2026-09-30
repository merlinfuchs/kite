import { numericOrPlaceholder } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
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
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder("ID of the channel to send the message to."),
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { schema: nodeActionMessageCreateResultSchema },
  run: { kind: "custom" },
};
