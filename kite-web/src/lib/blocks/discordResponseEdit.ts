import { nodeActionResponseEditDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionResponseEditDataSchema,
  inputs: [
    "response_target",
    "message_template_id",
    "message_data",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionResponseEditResultSchema },
  run: { kind: "custom" },
};
