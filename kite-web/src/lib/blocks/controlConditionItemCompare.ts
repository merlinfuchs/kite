import { nodeConditionItemCompareDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionItemCompare: BlockDefinition = {
  type: "control_condition_item_compare",
  title: "Match Condition",
  description: "Run actions if the two values are equal.",
  icon: "circle-help",
  component: "condition_item",
  schema: nodeConditionItemCompareDataSchema,
  inputs: ["condition_item_compare_mode", "condition_item_compare_value"],
  run: { kind: "custom" },
};
