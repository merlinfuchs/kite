import { nodeActionAiWebSearchCompletionDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionAiWebSearchCompletionDataSchema,
  inputs: ["ai_web_search_data", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
