import { nodeActionAiChatCompletionDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionAiChatCompletionDataSchema,
  inputs: ["ai_chat_completion_data", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
