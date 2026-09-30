import { z } from "zod";
import {
  aiMaxCompletionTokensSchema,
  aiModelSchema,
  templated,
} from "../flow/dataSchema";
import { getAiModelCredits } from "../flow/aiModels";
import { BlockDefinition } from "./types";

export const aiChatCompletion: BlockDefinition = {
  type: "action_ai_chat_completion",
  title: "Ask AI",
  description:
    "Ask artificial intelligence a question or let it respond to a prompt",
  icon: "brain-circuit",
  category: "AI",
  credits: (data) =>
    getAiModelCredits(data.ai_chat_completion_data?.model, "chat"),
  fields: [
    {
      name: "ai_chat_completion_data",
      type: "string",
      schema: z
        .object({
          model: aiModelSchema,
          system_prompt: templated(
            z.string().max(2000),
            "Instructions for how the AI should behave."
          ).optional(),
          prompt: templated(
            z.string().max(2000).min(1),
            "Message the AI responds to."
          ),
          max_completion_tokens: aiMaxCompletionTokensSchema,
        })
        .describe("The prompt and model settings."),
    },
  ],
  result: { schema: z.string().describe("The answer of the AI.") },
  run: { kind: "custom" },
};
