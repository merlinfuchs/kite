import { nodeConditionItemUserDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionItemUser: BlockDefinition = {
  type: "control_condition_item_user",
  title: "Match User",
  description: "Run actions if the user meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  schema: nodeConditionItemUserDataSchema,
  inputs: ["condition_item_user_mode", "condition_item_user_value"],
  run: { kind: "custom" },
};
