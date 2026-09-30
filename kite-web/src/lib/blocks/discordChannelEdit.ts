import { channelDataSchema, channelTargetSchema } from "../flow/dataSchema";
import { nodeActionChannelEditResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordChannelEdit: BlockDefinition = {
  type: "action_channel_edit",
  title: "Edit channel",
  description: "Edit a channel or thread",
  icon: "folder-pen",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "channel_target",
      type: "snowflake",
      schema: channelTargetSchema,
    },
    {
      name: "channel_data",
      type: "string",
      schema: channelDataSchema,
    },
  ],
  audit_log_reason: true,
  result: { schema: nodeActionChannelEditResultSchema },
  run: { kind: "custom" },
};
