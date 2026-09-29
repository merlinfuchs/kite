import { nodeConditionRoleDataSchema } from "../flow/dataSchema";
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
  schema: nodeConditionRoleDataSchema,
  inputs: [
    "condition_role_base_value",
    "condition_allow_multiple",
    "custom_label",
  ],
  run: { kind: "custom" },
};
