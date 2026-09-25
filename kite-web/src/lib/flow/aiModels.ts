import { AIModelLarge, AIModelMedium, AIModelSmall } from "../types/flow.gen";

// AI blocks store a tier, not a model, so the service can point a tier at a
// different model without touching stored flows. Mirrors aiModelTiers in
// kite-service/pkg/flow/data.go; update the model names here when it changes.
export const aiModelTiers = [
  {
    value: AIModelSmall,
    label: "Fast",
    model: "gpt-6-luna",
    credits: { chat: 5, search: 25 },
  },
  {
    value: AIModelMedium,
    label: "Balanced",
    model: "gpt-6-luna with reasoning",
    credits: { chat: 20, search: 100 },
  },
  {
    value: AIModelLarge,
    label: "Smartest",
    model: "gpt-6-sol",
    credits: { chat: 100, search: 500 },
  },
] as const;

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
