import { nodeActionResponseCreateDataSchema } from "../flow/dataSchema";
import { nodeActionResponseCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordResponseCreate: BlockDefinition = {
  type: "action_response_create",
  title: "Create response message",
  description: "Bot replies to the interaction with a message",
  icon: "message-circle-reply",
  category: "Responses",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionResponseCreateDataSchema,
  inputs: [
    "message_template_id",
    "message_data",
    "message_ephemeral",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionResponseCreateResultSchema },
  run: { kind: "custom" },
};
