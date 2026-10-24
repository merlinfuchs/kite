import { idConditionModeSchema } from "../flow/dataSchema";
import { conditionItemFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemCategory: BlockDefinition = {
  type: "control_condition_item_category",
  id_prefix: "condition",
  title: "Match Category",
  description: "Run actions if the channel is in the category.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: conditionItemFields(
    "category",
    idConditionModeSchema,
    "ID of the category the channel has to be in, or not be in."
  ),
  run: { kind: "custom" },
};
