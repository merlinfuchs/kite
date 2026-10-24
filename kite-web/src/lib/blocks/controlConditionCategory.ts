import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { conditionFields } from "./fields";
import { BlockDefinition } from "./types";

// Each branch checks the category of the channel, so one branch covers every
// channel in a category.
export const controlConditionCategory: BlockDefinition = {
  type: "control_condition_category",
  id_prefix: "condition",
  title: "Category Condition",
  description: "Run actions based on the category of a channel.",
  icon: "folder-tree",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_category"],
  component: "condition",
  fields: conditionFields(
    "category",
    "ID of the channel whose category each branch checks. Empty for the channel the flow runs in. A thread is in the category of its channel, and a category is in itself."
  ).map((field) =>
    field.name === "condition_base_value"
      ? {
          ...field,
          schema: templated(
            z.string(),
            "ID of the channel whose category each branch checks. Empty for the channel the flow runs in. A thread is in the category of its channel, and a category is in itself."
          ).optional(),
        }
      : field
  ),
  run: { kind: "custom" },
};
