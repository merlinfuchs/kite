import { describe, expect, it } from "vitest";
import { getNodeCreditsCost, getNodeValues } from "./nodes";

const aiChatCompletionSchema = getNodeValues("action_ai_chat_completion")
  .dataSchema!;

function aiData(model?: string) {
  return { ai_chat_completion_data: { model, prompt: "hi" } };
}

function credits(type: string, model?: string) {
  return getNodeCreditsCost(getNodeValues(type), aiData(model));
}

describe("AI model tiers", () => {
  it("accepts tiers and models stored before tiers existed", () => {
    for (const model of [
      undefined,
      "",
      "small",
      "medium",
      "large",
      "gpt-4.1",
      "gpt-4.1-mini",
      "gpt-4.1-nano",
      "gpt-5-nano",
      "gpt-5.4-nano",
      "gpt-4o-mini",
    ]) {
      expect(
        aiChatCompletionSchema.safeParse(aiData(model)).success,
        String(model)
      ).toBe(true);
    }
  });

  it("rejects other models", () => {
    for (const model of ["gpt-6-luna", "o3", "toString"]) {
      expect(
        aiChatCompletionSchema.safeParse(aiData(model)).success,
        model
      ).toBe(false);
    }
  });

  it("prices old models like the tier they now run as", () => {
    const cases: [string | undefined, number, number][] = [
      [undefined, 5, 250],
      ["small", 5, 250],
      ["gpt-4o-mini", 5, 250],
      ["gpt-5-nano", 5, 250],
      ["gpt-5.4-nano", 5, 250],
      ["medium", 20, 300],
      ["gpt-4.1-mini", 20, 300],
      ["large", 100, 600],
      ["gpt-4.1", 100, 600],
      ["unknown", 100, 600],
    ];
    for (const [model, chat, search] of cases) {
      expect(credits("action_ai_chat_completion", model), model).toBe(chat);
      expect(credits("action_ai_web_search", model), model).toBe(search);
    }
  });
});
