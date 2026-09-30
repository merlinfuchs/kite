import { responseTargetSchema } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { nodeActionResponseEditResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordResponseEdit: BlockDefinition = {
  type: "action_response_edit",
  title: "Edit response message",
  description: "Bot edits an existing interaction response message",
  icon: "pen",
  category: "Responses",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "message_target",
      type: "string",
      input: "response_target",
      schema: responseTargetSchema,
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { schema: nodeActionResponseEditResultSchema },
  run: { kind: "custom" },
};
