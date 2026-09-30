import { channelDataSchema, guildTargetSchema } from "../flow/dataSchema";
import { nodeActionChannelCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordChannelCreate: BlockDefinition = {
  type: "action_channel_create",
  title: "Create channel",
  description: "Create a channel",
  icon: "folder-plus",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "guild_target",
      type: "snowflake",
      schema: guildTargetSchema.optional(),
    },
    {
      name: "channel_data",
      type: "string",
      schema: channelDataSchema,
    },
  ],
  audit_log_reason: true,
  result: { schema: nodeActionChannelCreateResultSchema },
  run: { kind: "custom" },
};
