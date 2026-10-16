import { z } from "zod";
import { BlockDefinition } from "./types";

export const entryFunction: BlockDefinition = {
  type: "entry_function",
  title: "Function",
  description:
    "Runs when a Call function block calls the function. Drop different actions here!",
  icon: "square-function",
  contexts: ["function"],
  fixed: true,
  component: "entry_function",
  custom_label: false,
  strict_settings: true,
  fields: [
    {
      name: "name",
      schema: z
        .string()
        .max(100)
        .min(1)
        .describe("Name of the function, shown when picking it to call."),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .optional()
        .describe("Description of what the function does."),
    },
    {
      name: "function_parameters",
      schema: z
        .array(
          z.object({
            name: z
              .string()
              .max(32)
              .min(1)
              .regex(
                /^[a-z0-9_]+$/,
                "Must be only lowercase alphanumeric characters and underscores"
              )
              .describe(
                "Name of the parameter. Its value can be read with {{arg('name')}}."
              ),
            description: z
              .string()
              .max(100)
              .optional()
              .describe("What to pass for the parameter."),
          })
        )
        .max(25)
        .optional()
        .describe(
          "Values the function receives from the Call function block, read with {{arg('name')}}."
        ),
    },
  ],
  run: { kind: "custom" },
};
