import { idConditionModeSchema } from "../flow/dataSchema";
import { conditionItemFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemRole: BlockDefinition = {
  type: "control_condition_item_role",
  title: "Match Role",
  description: "Run actions if the role meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: conditionItemFields(
    "role",
    idConditionModeSchema,
    "ID to compare the base value with."
  ),
  run: { kind: "custom" },
};
