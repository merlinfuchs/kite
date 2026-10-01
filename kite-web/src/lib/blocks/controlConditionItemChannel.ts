import { idConditionModeSchema } from "../flow/dataSchema";
import { conditionItemFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemChannel: BlockDefinition = {
  type: "control_condition_item_channel",
  title: "Match Channel",
  description: "Run actions if the channel meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: conditionItemFields(
    "channel",
    idConditionModeSchema,
    "ID to compare the base value with."
  ),
  run: { kind: "custom" },
};
