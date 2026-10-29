import { z } from "zod";
import { placeholderRegex, templated } from "../flow/dataSchema";
import { errorColor } from "./colors";
import { BlockDefinition } from "./types";

const conditionMessage =
  "Must be a single {{ }} placeholder, like {{result('count') > 5}}";

export const controlIf: BlockDefinition = {
  type: "control_if",
  id_prefix: "if",
  title: "If",
  description: "Run actions if something is true, and others if it isn't.",
  icon: "split",
  category: "Conditions",
  outputs: ["yes", "no"],
  output_labels: {
    yes: { label: "Yes" },
    no: { label: "No", color: errorColor },
  },
  component: "branches",
  strict_settings: true,
  fields: [
    {
      name: "condition_base_value",
      input: "condition_compare_base_value",
      schema: templated(
        z
          .string()
          .regex(placeholderRegex, conditionMessage)
          .or(z.string().regex(/^(true|false)$/, conditionMessage)),
        "Placeholder that is true or false, like {{result('count') > 5}} or {{message.content != ''}}. The yes branch runs if it's true, yes or 1, in any case, and the no branch otherwise. For more than two branches, use control_condition_compare."
      ),
    },
  ],
  run: { kind: "custom" },
};
