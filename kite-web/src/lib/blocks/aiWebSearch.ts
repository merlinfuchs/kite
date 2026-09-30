import { z } from "zod";
import {
  aiMaxCompletionTokensSchema,
  aiModelSchema,
  templated,
} from "../flow/dataSchema";
import { getAiModelCredits } from "../flow/aiModels";
import { BlockDefinition } from "./types";

export const aiWebSearch: BlockDefinition = {
  type: "action_ai_web_search",
  title: "Search the Web",
  description: "Search the web for the latest information using AI",
  icon: "search",
  category: "AI",
  credits: (data) =>
    getAiModelCredits(data.ai_chat_completion_data?.model, "search"),
  fields: [
    {
      name: "ai_chat_completion_data",
      type: "string",
      input: "ai_web_search_data",
      schema: z
        .object({
          model: aiModelSchema,
          prompt: templated(
            z.string().max(2000).min(1),
            "What to search the web for."
          ),
          max_completion_tokens: aiMaxCompletionTokensSchema,
        })
        .describe("The search query and model settings."),
    },
  ],
  result: {
    schema: z
      .string()
      .describe("The answer of the AI, based on what it found."),
  },
  run: { kind: "custom" },
};
