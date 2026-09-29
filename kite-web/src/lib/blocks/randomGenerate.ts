import { nodeActionRandomGenerateDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const randomGenerate: BlockDefinition = {
  type: "action_random_generate",
  title: "Generate Random Number",
  description: "Generate a random number in a range",
  icon: "dices",
  category: "Utilities",
  credits: 1,
  schema: nodeActionRandomGenerateDataSchema,
  inputs: ["random_min", "random_max", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
