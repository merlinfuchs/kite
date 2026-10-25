import { AnyZodObject, z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const randomModes = [
  { value: "number", label: "Number in range" },
  { value: "item", label: "Item from list" },
] as const;

export const randomGenerate: BlockDefinition = {
  type: "action_random_generate",
  id_prefix: "random",
  title: "Generate Random Value",
  description: "Generate a random number or pick a random item from a list",
  icon: "dices",
  category: "Utilities",
  credits: 1,
  fields: [
    {
      name: "random_mode",
      input: "random",
      schema: z
        .enum(["number", "item"])
        .optional()
        .describe(
          "number generates a number from random_min to below random_max, item picks one of random_items. Defaults to number."
        ),
    },
    {
      name: "random_min",
      input: "random",
      schema: numericOrPlaceholder(
        "Smallest number that can be generated."
      ).optional(),
    },
    {
      name: "random_max",
      input: "random",
      schema: numericOrPlaceholder(
        "Upper bound of the generated number. The number is always below it."
      ).optional(),
    },
    {
      name: "random_items",
      input: "random",
      schema: templated(
        z.string().max(10000).optional(),
        "Items to pick from, one per line, or a single placeholder of a list like {{var('prizes')}}, whose items keep their type."
      ),
    },
  ],
  refine: (schema: AnyZodObject) =>
    schema.superRefine((data, ctx) => {
      const required =
        data.random_mode === "item"
          ? ["random_items"]
          : ["random_min", "random_max"];
      for (const field of required) {
        if (!data[field]?.trim()) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Required",
            path: [field],
          });
        }
      }
    }),
  result: {
    schema: z.unknown().describe("The generated number, or the picked item."),
  },
  run: { kind: "custom" },
};
