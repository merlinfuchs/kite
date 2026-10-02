import { comparisonModeSchema } from "../flow/dataSchema";
import { conditionItemFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemCompare: BlockDefinition = {
  type: "control_condition_item_compare",
  title: "Match Condition",
  description: "Run actions if the two values are equal.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: conditionItemFields(
    "compare",
    comparisonModeSchema.describe(
      "How the condition's base value is compared to this branch's value."
    ),
    "Value to compare the base value with."
  ),
  run: { kind: "custom" },
};
