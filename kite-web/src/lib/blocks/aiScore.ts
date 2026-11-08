import { z } from "zod";
import {
  aiDecisionInputSchema,
  aiDecisionQuestionSchema,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const aiScoreMaxLevels = 10;

export const aiScore: BlockDefinition = {
  type: "action_ai_score",
  id_prefix: "score",
  title: "AI Score",
  description: "Let AI rate something on a scale you define",
  icon: "gauge",
  category: "AI",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "ai_decision_data",
      input: "ai_score_data",
      schema: z
        .object({
          question: aiDecisionQuestionSchema,
          input: aiDecisionInputSchema,
          levels: z
            .array(z.string().min(1).max(200))
            .min(2)
            .max(aiScoreMaxLevels)
            .describe(
              "What each level of the scale means, lowest first, like ['Not urgent', 'Urgent', 'Very urgent']."
            ),
        })
        .describe("The question, what it's about and the scale."),
    },
  ],
  result: {
    schema: z
      .number()
      .describe(
        "The level, from 1 for the first to the number of levels. It can be in between, like 2.4."
      ),
  },
  run: { kind: "custom" },
};
