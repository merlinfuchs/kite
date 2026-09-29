import { nodeActionChannelCreateDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionChannelCreateDataSchema,
  inputs: [
    "guild_target",
    "channel_data",
    "audit_log_reason",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionChannelCreateResultSchema },
  run: { kind: "custom" },
};
