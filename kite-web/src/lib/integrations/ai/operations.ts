import { z } from "zod";
import {
  aiTranslate,
  aiTranslateMaxLength,
  aiTranslateStepLength,
} from "../../blocks/aiTranslate";
import { aiScore, aiScoreMaxLevels } from "../../blocks/aiScore";
import { controlAiCheck } from "../../blocks/controlAiCheck";
import { controlConditionAiChoice } from "../../blocks/controlConditionAiChoice";
import {
  aiAnswerLimits,
  aiInputStepLength,
  aiModelTiers,
  aiModelTierValues,
} from "../../flow/aiModels";
import { translateLanguageCodes } from "../../flow/translateLanguages";
import { formatNumber as n } from "../../utils";
import { HandlerOperationDefinition } from "../handlerOperations";

// The AI blocks for code, which cost the same. Search the Web isn't one, as
// its searches take longer than the 10 seconds a run of code has. Run by
// kite-service/pkg/flow/operation_handlers_ai.go.

// Costs 5 credits with small, 20 with medium, and 100 with large.
const tierCredits = `Costs ${new Intl.ListFormat("en").format(
  aiModelTiers.map(
    (t, i) => `${n(t.credits.chat)}${i === 0 ? " credits" : ""} with ${t.value}`
  )
)}`;

const question = z.string().describe("The question about the input.");
const input = z
  .string()
  .describe("What the question is about, like the text of a message.");

export const aiOperations: HandlerOperationDefinition[] = [
  {
    id: "ask",
    description: `Ask AI a question or let it respond to a prompt, like the Ask AI block, and get its answer. ${tierCredits}.`,
    credits: Math.min(...aiModelTiers.map((t) => t.credits.chat)),
    variable_credits: true,
    params: z
      .object({
        prompt: z
          .string()
          .describe(
            `Message the AI responds to. Up to ${n(
              Math.max(...aiModelTiers.map((t) => t.maxInputLength))
            )} bytes together with the system prompt, or ${n(
              Math.min(...aiModelTiers.map((t) => t.maxInputLength))
            )} with large. Each ${n(
              aiInputStepLength
            )} bytes after the first cost the model's credits again.`
          ),
        system_prompt: z
          .string()
          .optional()
          .describe("Instructions for how the AI should behave."),
        model: z
          .enum(aiModelTierValues)
          .optional()
          .describe(
            "small is fast, medium reasons first, large is the smartest. Defaults to small."
          ),
        max_tokens: z
          .number()
          .optional()
          .describe(`The longest answer. ${aiAnswerLimits}`),
      })
      .strict(),
    result: z.string().describe("The answer."),
  },
  {
    id: "check",
    description:
      "Let AI answer a yes or no question about the input, like the AI Check block. Compare the result with a threshold, like 0.5.",
    credits: controlAiCheck.credits as number,
    params: z.object({ question, input }).strict(),
    result: z.number().describe("The probability of yes, from 0 to 1."),
  },
  {
    id: "choose",
    description:
      "Let AI pick the option that fits the input best, like the AI Choice block.",
    credits: controlConditionAiChoice.credits as number,
    params: z
      .object({
        question,
        input,
        options: z
          .record(z.string(), z.string())
          .describe(
            'The options by name, each with a description of when it fits, or "" if the name says enough.'
          ),
      })
      .strict(),
    result: z.object({
      choice: z
        .string()
        .nullable()
        .describe("The name of the option, or null if none fits."),
      confidence: z
        .number()
        .describe("How sure the AI is about the choice, from 0 to 1."),
    }),
  },
  {
    id: "score",
    description:
      "Let AI rate the input on a scale of levels, like the AI Score block.",
    credits: aiScore.credits as number,
    params: z
      .object({
        question,
        input,
        levels: z
          .array(z.string())
          .describe(
            `2 to ${aiScoreMaxLevels} descriptions of the levels, from the lowest to the highest.`
          ),
      })
      .strict(),
    result: z.number().describe("The level, from 1."),
  },
  {
    id: "translate",
    description: `Let AI translate text to another language, like the AI Translate block. Up to ${n(
      aiTranslateMaxLength
    )} characters. Costs 1 credit for each ${n(
      aiTranslateStepLength
    )} characters started.`,
    credits: aiTranslate.credits as number,
    variable_credits: true,
    params: z
      .object({
        text: z.string().describe("The text to translate."),
        language: z
          .enum(translateLanguageCodes)
          .describe("The language to translate to, like es or pt-BR."),
      })
      .strict(),
    result: z.object({
      text: z.string().describe("The translated text."),
      source_language: z
        .string()
        .describe('The code of the language the text was in, or "other".'),
    }),
  },
];
