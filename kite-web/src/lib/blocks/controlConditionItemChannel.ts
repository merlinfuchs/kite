import { nodeConditionItemIdDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionItemChannel: BlockDefinition = {
  type: "control_condition_item_channel",
  title: "Match Channel",
  description: "Run actions if the channel meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  schema: nodeConditionItemIdDataSchema,
  inputs: ["condition_item_channel_mode", "condition_item_channel_value"],
  run: { kind: "custom" },
};
