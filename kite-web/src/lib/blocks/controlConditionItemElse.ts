import { nodeEmptyDataSchema } from "../flow/dataSchema";
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
  schema: nodeEmptyDataSchema,
  inputs: [],
  run: { kind: "custom" },
};
