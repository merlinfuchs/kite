import { responseTargetSchema } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
  responseTargetSetting,
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
  fields: [responseTargetSetting, messageTemplateField, messageDataField],
  refine: requireMessage,
  result: { schema: nodeActionResponseEditResultSchema },
  run: { kind: "custom" },
};
