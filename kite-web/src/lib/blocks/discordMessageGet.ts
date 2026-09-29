import { nodeActionMessageGetDataSchema } from "../flow/dataSchema";
import { nodeActionMessageGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordMessageGet: BlockDefinition = {
  type: "action_message_get",
  title: "Get channel message",
  description: "Get a message from a channel",
  icon: "mail-search",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageGetDataSchema,
  inputs: ["message_target", "temporary_name", "custom_label"],
  result: { schema: nodeActionMessageGetResultSchema },
  run: { kind: "custom" },
};
