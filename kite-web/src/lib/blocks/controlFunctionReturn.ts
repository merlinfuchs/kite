import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlFunctionReturn: BlockDefinition = {
  type: "control_function_return",
  title: "Return",
  description:
    "End the function and return a value to the Call function block that called it.",
  icon: "corner-down-left",
  category: "Functions",
  contexts: ["function"],
  outputs: [],
  component: "control_function_return",
  strict_settings: true,
  fields: [
    {
      name: "function_return_value",
      schema: templated(
        z.string(),
        "Value to return, which is the result of the Call function block. Leave empty to return nothing."
      ).optional(),
    },
    {
      name: "function_return_stop",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Also stop the flow that called the function, so nothing after the Call function block runs. Useful for checks like permissions that already responded to the user."
        ),
    },
  ],
  run: { kind: "custom" },
};
