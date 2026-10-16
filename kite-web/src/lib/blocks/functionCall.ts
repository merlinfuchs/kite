import { z } from "zod";
import { templated, userPicked } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const functionCall: BlockDefinition = {
  type: "action_function_call",
  title: "Call function",
  description:
    "Run one of your functions and continue with the value it returns.",
  icon: "square-function",
  category: "Functions",
  credits: 0,
  strict_settings: true,
  fields: [
    {
      name: "function_id",
      schema: userPicked(z.string(), "ID of an existing function of the app."),
    },
    {
      name: "function_arguments",
      schema: z
        .array(
          z.object({
            name: z
              .string()
              .min(1)
              .describe("Name of the function's parameter."),
            value: templated(
              z.string(),
              "Value passed for the parameter."
            ).optional(),
          })
        )
        .max(25)
        .optional()
        .describe("The values passed for the function's parameters."),
    },
  ],
  result: {
    schema: z
      .unknown()
      .describe(
        "The value the function returned with its Return block, or nothing if it didn't."
      ),
  },
  run: { kind: "custom" },
};
