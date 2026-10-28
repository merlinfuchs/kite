import { AIModelLarge, AIModelMedium, AIModelSmall } from "../types/flow.gen";

// AI blocks store a tier, not a model, so the service can point a tier at a
// different model without touching stored flows. Mirrors aiModelTiers in
// kite-service/pkg/flow/data.go; update the model names here when it changes.
export const aiModelTiers = [
  {
    value: AIModelSmall,
    label: "Fast",
    model: "gpt-6-luna",
    maxInputLength: 64000,
    credits: { chat: 5, search: 25 },
  },
  {
    value: AIModelMedium,
    label: "Balanced",
    model: "gpt-6-luna with reasoning",
    maxInputLength: 64000,
    credits: { chat: 20, search: 100 },
  },
  {
    value: AIModelLarge,
    label: "Smartest",
    model: "gpt-6-sol",
    maxInputLength: 8000,
    credits: { chat: 100, search: 500 },
  },
] as const;

// A block's credits cover a step of prompt and one of answer. Every further
// step costs the tier's chat credits again. Mirrors aiInputStepLength and the
// output tokens in kite-service/pkg/flow/data.go.
export const aiInputStepLength = 8000;
export const aiOutputStepTokens = 500;
export const aiMaxOutputTokens = 1000;

const n = (v: number) => v.toLocaleString("en-US");
const smartest = aiModelTiers.find((t) => t.value === AIModelLarge)!;

// For the descriptions of the prompt and answer length of AI blocks.
export const aiPromptLimits = `Up to ${n(
  aiModelTiers[0].maxInputLength
)} bytes with placeholders filled in, or ${n(
  smartest.maxInputLength
)} with the ${smartest.label} model. Each ${n(
  aiInputStepLength
)} bytes after the first cost the model's Ask AI credits again.`;
export const aiAnswerLimits = `Up to ${n(
  aiMaxOutputTokens
)} tokens, including reasoning, and ${n(
  aiOutputStepTokens
)} if unset. Answers over ${n(
  aiOutputStepTokens
)} tokens cost the model's Ask AI credits once more.`;

type AiModelTier = (typeof aiModelTiers)[number]["value"];

export const aiModelTierValues = aiModelTiers.map((t) => t.value) as [
  AiModelTier,
  ...AiModelTier[]
];

// What flows stored before tiers existed. A block keeps its old value until
// someone edits it. The editor saved gpt-5.4-nano for its "gpt-5-nano" option
// from May to August 2026.
const legacyAiModels = new Map<unknown, AiModelTier>([
  ["gpt-4o-mini", AIModelSmall],
  ["gpt-4.1-nano", AIModelSmall],
  ["gpt-5-nano", AIModelSmall],
  ["gpt-5.4-nano", AIModelSmall],
  ["gpt-4.1-mini", AIModelMedium],
  ["gpt-4.1", AIModelLarge],
]);

export function resolveAiModel(model: unknown): unknown {
  if (model === undefined || model === "") return undefined;
  return legacyAiModels.get(model) ?? model;
}

export function getAiModelTier(model: unknown) {
  const value = resolveAiModel(model) ?? AIModelSmall;
  return aiModelTiers.find((t) => t.value === value);
}

// Unknown models are priced at the ceiling of each field, as the service does.
const maxAiModelCredits = {
  chat: Math.max(...aiModelTiers.map((t) => t.credits.chat)),
  search: Math.max(...aiModelTiers.map((t) => t.credits.search)),
};

export function getAiModelCredits(model: unknown, kind: "chat" | "search") {
  return (getAiModelTier(model)?.credits ?? maxAiModelCredits)[kind];
}
