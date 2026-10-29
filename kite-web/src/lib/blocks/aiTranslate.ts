import { z } from "zod";
import { placeholderRegex, templated } from "../flow/dataSchema";
import { translateLanguageCodes } from "../flow/translateLanguages";
import { BlockDefinition } from "./types";

// Matches aiTranslateMaxLength and aiTranslateStepLength in
// kite-service/pkg/flow/handlers_ai.go.
export const aiTranslateMaxLength = 4000;
export const aiTranslateStepLength = 1000;

const languageMessage =
  "Must be one of the languages, or a single {{ }} placeholder";

export const aiTranslate: BlockDefinition = {
  type: "action_ai_translate",
  id_prefix: "translate",
  title: "AI Translate",
  description: "Let AI translate text to another language",
  icon: "languages",
  category: "AI",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "ai_translate_data",
      schema: z
        .object({
          text: templated(
            z.string().min(1).max(aiTranslateMaxLength),
            `What to translate, like {{message.content}}. Up to ${aiTranslateMaxLength.toLocaleString(
              "en-US"
            )} characters with placeholders filled in. Each ${aiTranslateStepLength.toLocaleString(
              "en-US"
            )} characters after the first cost 1 more credit.`
          ),
          language: z
            .enum(translateLanguageCodes, { message: languageMessage })
            .or(z.string().regex(placeholderRegex, languageMessage))
            .describe(
              "The language to translate to, or a single placeholder that becomes its code, like es or pt-BR."
            ),
        })
        .describe("The text and the language to translate it to."),
    },
  ],
  result: {
    schema: z
      .object({
        text: z.string().describe("The translated text"),
        source_language: z
          .string()
          .describe(
            "Code of the language the text was in, like es, or other if it's none of the languages the block translates to"
          ),
      })
      .describe("The translation"),
  },
  run: { kind: "custom" },
};
