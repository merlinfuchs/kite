import { conditionFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionRole: BlockDefinition = {
  type: "control_condition_role",
  title: "Role Condition",
  description: "Run actions based on a role.",
  icon: "bookmark",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_role"],
  component: "condition_role",
  fields: conditionFields("role", "ID of the role that each branch checks."),
  run: { kind: "custom" },
};
