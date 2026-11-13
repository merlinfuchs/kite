import { z } from "zod";
import {
  aiDecisionInputSchema,
  aiDecisionQuestionSchema,
  aiProbabilitySchema,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlConditionAiChoice: BlockDefinition = {
  type: "control_condition_ai_choice",
  id_prefix: "choice",
  title: "AI Choice",
  description: "Let AI pick one of several options and run its actions.",
  icon: "split",
  category: "Conditions",
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_ai_choice"],
  continues_from: "control_condition_item_ai_choice",
  component: "condition",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "ai_decision_data",
      input: "ai_choice_data",
      schema: z
        .object({
          question: aiDecisionQuestionSchema,
          input: aiDecisionInputSchema,
          min_confidence: aiProbabilitySchema(
            "How confident the AI has to be, from 0 to 1. Below it, the else branch runs instead. Defaults to 0."
          ),
        })
        .describe("The question and what it's about."),
    },
  ],
  result: {
    schema: z
      .string()
      .describe(
        "The name of the option the AI picked, also when it wasn't confident enough."
      ),
  },
  run: { kind: "custom" },
};
