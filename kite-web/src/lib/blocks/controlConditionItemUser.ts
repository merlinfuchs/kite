import { z } from "zod";
import { conditionItemFields } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemUser: BlockDefinition = {
  type: "control_condition_item_user",
  title: "Match User",
  description: "Run actions if the user meets the criteria.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: conditionItemFields(
    "user",
    z
      .enum([
        "equal",
        "not_equal",
        "has_role",
        "not_has_role",
        "has_permission",
        "not_has_permission",
      ])
      .describe("What to check about the user."),
    "User ID for equal and not_equal, role ID for has_role and not_has_role, or permission bitfield for has_permission and not_has_permission."
  ),
  run: { kind: "custom" },
};
