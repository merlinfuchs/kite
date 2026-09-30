import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { variableSettings } from "./fields";
import { BlockDefinition } from "./types";

export const variableSet: BlockDefinition = {
  type: "action_variable_set",
  title: "Set stored variable",
  description: "Set the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  fields: [
    ...variableSettings,
    {
      name: "variable_operation",
      schema: z
        .enum(["overwrite", "append", "prepend", "increment", "decrement"])
        .describe(
          "How the value is combined with the stored one. increment and decrement add or subtract a number."
        ),
    },
    {
      name: "variable_value",
      schema: templated(z.string(), "Value to store."),
    },
  ],
  result: { schema: z.unknown().describe("The new value of the variable.") },
  run: { kind: "custom" },
};
