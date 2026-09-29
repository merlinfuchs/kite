import { nodeActionPollCreateDataSchema } from "../flow/dataSchema";
import { nodeActionPollCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordPollCreate: BlockDefinition = {
  type: "action_poll_create",
  title: "Create poll",
  description: "Bot sends a poll to a channel",
  icon: "vote",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionPollCreateDataSchema,
  inputs: ["channel_target", "poll_data", "temporary_name", "custom_label"],
  result: { schema: nodeActionPollCreateResultSchema },
  run: { kind: "custom" },
};
