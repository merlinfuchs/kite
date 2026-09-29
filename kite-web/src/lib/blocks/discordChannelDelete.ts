import { nodeActionChannelDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordChannelDelete: BlockDefinition = {
  type: "action_channel_delete",
  title: "Delete channel",
  description: "Delete a channel or thread ",
  icon: "folder-x",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionChannelDeleteDataSchema,
  inputs: ["channel_target", "audit_log_reason", "custom_label"],
  run: { kind: "custom" },
};
