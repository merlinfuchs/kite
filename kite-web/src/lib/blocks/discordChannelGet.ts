import { channelTargetSchema } from "../flow/dataSchema";
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
  allow_unknown_settings: true,
  fields: [
    { name: "channel_target", type: "snowflake", schema: channelTargetSchema },
  ],
  result: { schema: nodeActionChannelGetResultSchema },
  run: { kind: "custom" },
};
