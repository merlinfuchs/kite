import { nodeConditionCompareDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionCompare: BlockDefinition = {
  type: "control_condition_compare",
  title: "Comparison Condition",
  description: "Run actions based on the difference between two values.",
  icon: "arrow-left-right",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_compare"],
  component: "condition_compare",
  schema: nodeConditionCompareDataSchema,
  inputs: [
    "condition_compare_base_value",
    "condition_allow_multiple",
    "custom_label",
  ],
  run: { kind: "custom" },
};
