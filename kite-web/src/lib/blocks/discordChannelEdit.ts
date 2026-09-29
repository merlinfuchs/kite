import { nodeActionChannelEditDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionChannelEditDataSchema,
  inputs: [
    "channel_target",
    "channel_data",
    "audit_log_reason",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionChannelEditResultSchema },
  run: { kind: "custom" },
};
