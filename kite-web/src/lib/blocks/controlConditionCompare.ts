import { conditionFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionCompare: BlockDefinition = {
  type: "control_condition_compare",
  id_prefix: "condition",
  title: "Comparison Condition",
  description: "Run actions based on the difference between two values.",
  icon: "arrow-left-right",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_compare"],
  continues_from: "control_condition_item_compare",
  component: "condition",
  fields: conditionFields(
    "compare",
    "Value that each branch compares against."
  ),
  run: { kind: "custom" },
};
