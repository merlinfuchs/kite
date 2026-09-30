import { messageTargetSchema, numericOrPlaceholder } from "../flow/dataSchema";
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
  fields: [
    {
      name: "channel_target",
      type: "snowflake",
      schema: numericOrPlaceholder(
        "ID of the channel the message is in. Defaults to the channel the flow runs in."
      ).optional(),
    },
    {
      name: "message_target",
      type: "snowflake",
      schema: messageTargetSchema,
    },
  ],
  result: { schema: nodeActionMessageGetResultSchema },
  run: { kind: "custom" },
};
