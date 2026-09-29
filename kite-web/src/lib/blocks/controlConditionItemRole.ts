import { nodeConditionItemIdDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionItemRole: BlockDefinition = {
  type: "control_condition_item_role",
  title: "Match Role",
  description: "Run actions if the role meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  schema: nodeConditionItemIdDataSchema,
  inputs: ["condition_item_role_mode", "condition_item_role_value"],
  run: { kind: "custom" },
};
