import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockDefinition, SettingsSchema } from "./types";

export const randomModes = [
  { value: "number", label: "Number in range" },
  { value: "item", label: "Item from list" },
  { value: "items", label: "Several items from list" },
] as const;

export const randomGenerate: BlockDefinition = {
  type: "action_random_generate",
  id_prefix: "random",
  title: "Generate Random Value",
  description: "Generate a random number or pick random items from a list",
  icon: "dices",
  category: "Utilities",
  credits: 1,
  fields: [
    {
      name: "random_mode",
      input: "random",
      schema: z
        .enum(["number", "item", "items"])
        .optional()
        .describe(
          "number generates a number from random_min to below random_max, item picks one of random_items, items picks random_count different ones of random_items as a list. Defaults to number."
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
    {
      name: "random_count",
      input: "random",
      schema: numericOrPlaceholder(
        "How many different items the items mode picks, from 1 to 100. It picks all of them, shuffled, if the list has fewer, and none from an empty list.",
        /^(100|[1-9][0-9]?)$/
      ).optional(),
    },
  ],
  refine: (schema: SettingsSchema) =>
    schema.superRefine((data, ctx) => {
      const required =
        data.random_mode === "item"
          ? ["random_items"]
          : data.random_mode === "items"
          ? ["random_items", "random_count"]
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
    schema: z
      .unknown()
      .describe(
        "The generated number, the picked item, or the list of picked items in random order."
      ),
  },
  run: { kind: "custom" },
};
