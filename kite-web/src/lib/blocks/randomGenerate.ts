import { z } from "zod";
import { numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const randomGenerate: BlockDefinition = {
  type: "action_random_generate",
  title: "Generate Random Number",
  description: "Generate a random number in a range",
  icon: "dices",
  category: "Utilities",
  credits: 1,
  fields: [
    {
      name: "random_min",
      type: "integer",
      schema: numericOrPlaceholder("Smallest number that can be generated."),
    },
    {
      name: "random_max",
      type: "integer",
      schema: numericOrPlaceholder(
        "Upper bound of the generated number. The number is always below it."
      ),
    },
  ],
  result: { schema: z.number().describe("The generated number.") },
  run: { kind: "custom" },
};
