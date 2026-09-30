import { responseTargetSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordResponseDelete: BlockDefinition = {
  type: "action_response_delete",
  title: "Delete response message",
  description: "Bot deletes an existing interaction response message",
  icon: "message-circle-x",
  category: "Responses",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "message_target",
      type: "string",
      input: "response_target",
      schema: responseTargetSchema,
    },
  ],
  run: { kind: "custom" },
};
