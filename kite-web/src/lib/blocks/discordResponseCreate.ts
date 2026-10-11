import { z } from "zod";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
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
  fields: [
    messageTemplateField,
    messageDataField,
    {
      name: "message_ephemeral",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether only the user who triggered the flow can see the response."
        ),
    },
  ],
  refine: requireMessage,
  result: { schema: nodeActionResponseCreateResultSchema },
  run: { kind: "custom" },
};
