import { nodeConditionUserDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionUser: BlockDefinition = {
  type: "control_condition_user",
  title: "User Condition",
  description: "Run actions based on a user.",
  icon: "user-search",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_user"],
  component: "condition_user",
  schema: nodeConditionUserDataSchema,
  inputs: [
    "condition_user_base_value",
    "condition_allow_multiple",
    "custom_label",
  ],
  run: { kind: "custom" },
};
