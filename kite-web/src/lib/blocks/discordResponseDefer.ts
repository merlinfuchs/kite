import { nodeActionResponseDeferDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordResponseDefer: BlockDefinition = {
  type: "action_response_defer",
  title: "Defer response",
  description:
    "Bot defers the response to the interaction to give time for further processing",
  icon: "message-circle-question",
  category: "Responses",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionResponseDeferDataSchema,
  inputs: ["message_ephemeral", "custom_label"],
  run: { kind: "custom" },
};
