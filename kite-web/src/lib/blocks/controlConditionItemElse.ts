import { errorColor } from "./colors";
import { BlockDefinition } from "./types";

export const controlConditionItemElse: BlockDefinition = {
  type: "control_condition_item_else",
  title: "Else",
  description: "Run actions if no other conditions are met.",
  icon: "circle-x",
  color: errorColor,
  fixed: true,
  component: "condition_item",
  custom_label: false,
  run: { kind: "custom" },
};
