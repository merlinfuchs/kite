import { z } from "zod";
import { comparisonModeSchema } from "../flow/dataSchema";
import { conditionItemFields, secretsOnlyWithEquality } from "./fields";
import { BlockDefinition } from "./types";

export const controlConditionItemCompare: BlockDefinition = {
  type: "control_condition_item_compare",
  id_prefix: "condition",
  title: "Match Condition",
  description: "Run actions if the two values are equal.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  fields: [
    ...conditionItemFields(
      "compare",
      comparisonModeSchema.describe(
        "How the condition's base value is compared to this branch's value. one_of, not_one_of and contains_any compare it with a list of values."
      ),
      "Value to compare the base value with. For one_of, not_one_of and contains_any, a list with one value per line, or a placeholder that holds a list."
    ).map((f) =>
      f.name === "condition_item_value" ? { ...f, secrets: true } : f
    ),
    {
      name: "condition_item_ignore_case",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether text is compared ignoring upper and lower case. Numbers are compared as they are."
        ),
    },
  ],
  refine: secretsOnlyWithEquality,
  run: { kind: "custom" },
};
