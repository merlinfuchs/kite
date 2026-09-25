// AI blocks store a tier, not a model, so the service can point a tier at a
// different model without touching stored flows. Mirrors aiModelTiers in
// kite-service/pkg/flow/data.go; update the model names here when it changes.
export const aiModelTiers = [
  {
    value: "small",
    label: "Fast",
    model: "gpt-6-luna",
    credits: { chat: 5, search: 25 },
  },
  {
    value: "medium",
    label: "Balanced",
    model: "gpt-6-luna with reasoning",
    credits: { chat: 20, search: 100 },
  },
  {
    value: "large",
    label: "Smartest",
    model: "gpt-6-sol",
    credits: { chat: 100, search: 500 },
  },
] as const;

export type AiModelTier = (typeof aiModelTiers)[number]["value"];

export const aiModelTierValues = aiModelTiers.map((t) => t.value) as [
  AiModelTier,
  ...AiModelTier[]
];

// What flows stored before tiers existed. A block keeps its old value until
// someone edits it.
const legacyAiModels = new Map<unknown, AiModelTier>([
  ["gpt-4o-mini", "small"],
  ["gpt-4.1-nano", "small"],
  ["gpt-5-nano", "small"],
  ["gpt-4.1-mini", "medium"],
  ["gpt-4.1", "large"],
]);

export function resolveAiModel(model: unknown): unknown {
  if (model === undefined || model === "") return undefined;
  return legacyAiModels.get(model) ?? model;
}

// Unknown models are priced like the most expensive tier, as the service does.
export function getAiModelTier(model: unknown) {
  const value = resolveAiModel(model) ?? "small";
  return (
    aiModelTiers.find((t) => t.value === value) ??
    aiModelTiers[aiModelTiers.length - 1]
  );
}
