import { z } from "zod";
import { BlockDefinition } from "./types";

export const controlConditionItemAiChoice: BlockDefinition = {
  type: "control_condition_item_ai_choice",
  id_prefix: "condition",
  title: "Option",
  description: "Run actions if the AI picks this option.",
  icon: "circle-help",
  component: "condition_item",
  custom_label: false,
  strict_settings: true,
  fields: [
    {
      name: "ai_choice_name",
      schema: z
        .string()
        .max(100)
        .regex(/\S/, "Required")
        .describe(
          "Name of the option, unique within the AI Choice. It's what the AI Choice block returns."
        ),
    },
    {
      name: "ai_choice_description",
      schema: z
        .string()
        .max(500)
        .optional()
        .describe("When to pick the option, if the name doesn't say."),
    },
  ],
  run: { kind: "custom" },
};
