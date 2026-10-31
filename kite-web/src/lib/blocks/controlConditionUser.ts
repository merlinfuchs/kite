import { conditionFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionUser: BlockDefinition = {
  type: "control_condition_user",
  id_prefix: "condition",
  title: "User Condition",
  description: "Run actions based on a user.",
  icon: "user-search",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_user"],
  continues_from: "control_condition_item_user",
  component: "condition",
  fields: conditionFields("user", "ID of the user that each branch checks."),
  run: { kind: "custom" },
};
