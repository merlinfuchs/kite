import { nodeConditionChannelDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionChannel: BlockDefinition = {
  type: "control_condition_channel",
  title: "Channel Condition",
  description: "Run actions based on a channel.",
  icon: "folder-search",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_channel"],
  component: "condition_channel",
  schema: nodeConditionChannelDataSchema,
  inputs: [
    "condition_channel_base_value",
    "condition_allow_multiple",
    "custom_label",
  ],
  run: { kind: "custom" },
};
