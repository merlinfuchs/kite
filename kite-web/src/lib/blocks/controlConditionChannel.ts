import { conditionFields } from "./fields";
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
  fields: conditionFields(
    "channel",
    "ID of the channel that each branch checks."
  ),
  run: { kind: "custom" },
};
