import { z } from "zod";
import {
  aiDecisionInputSchema,
  aiDecisionQuestionSchema,
  aiProbabilitySchema,
  decimalRegex,
} from "../flow/dataSchema";
import { errorColor } from "./colors";
import { BlockDefinition, SettingsSchema } from "./types";

export const controlAiCheck: BlockDefinition = {
  type: "control_ai_check",
  id_prefix: "check",
  title: "AI Check",
  description:
    "Let AI answer a yes or no question and run actions based on it.",
  icon: "circle-check-big",
  category: "Conditions",
  outputs: ["yes", "no", "unsure"],
  output_labels: {
    yes: { label: "Yes" },
    no: { label: "No", color: errorColor },
    unsure: { label: "Unsure" },
  },
  component: "branches",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "ai_decision_data",
      input: "ai_check_data",
      schema: z
        .object({
          question: aiDecisionQuestionSchema,
          input: aiDecisionInputSchema,
          yes_threshold: aiProbabilitySchema(
            "Probability of yes, from 0 to 1, from which the yes branch runs. Defaults to the no threshold, or 0.5."
          ),
          no_threshold: aiProbabilitySchema(
            "Probability of yes, from 0 to 1, below which the no branch runs. In between the two, the unsure branch runs. Defaults to the yes threshold, or 0.5."
          ),
        })
        .describe("The question and what it's about."),
    },
  ],
  refine: (schema: SettingsSchema) =>
    schema.superRefine((data, ctx) => {
      const yes = data.ai_decision_data?.yes_threshold;
      const no = data.ai_decision_data?.no_threshold;
      // An unset threshold is the other one. Placeholders are checked when
      // the flow runs.
      if (!yes || !no || !decimalRegex.test(yes) || !decimalRegex.test(no)) {
        return;
      }
      if (parseFloat(no) > parseFloat(yes)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Can't be above the yes threshold",
          path: ["ai_decision_data", "no_threshold"],
        });
      }
    }),
  result: {
    schema: z.number().describe("The probability of yes, from 0 to 1."),
  },
  run: { kind: "custom" },
};
