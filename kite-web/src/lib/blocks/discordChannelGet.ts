import { nodeActionChannelGetDataSchema } from "../flow/dataSchema";
import { nodeActionChannelGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordChannelGet: BlockDefinition = {
  type: "action_channel_get",
  title: "Get channel",
  description: "Get a channel by ID",
  icon: "folder-search",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionChannelGetDataSchema,
  inputs: ["channel_target", "temporary_name", "custom_label"],
  result: { schema: nodeActionChannelGetResultSchema },
  run: { kind: "custom" },
};
