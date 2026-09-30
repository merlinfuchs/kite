import { z } from "zod";
import {
  templated,
  variableIdSchema,
  variableScopeSchema,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableSet: BlockDefinition = {
  type: "action_variable_set",
  title: "Set stored variable",
  description: "Set the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  fields: [
    {
      name: "variable_id",
      type: "string",
      schema: variableIdSchema,
    },
    {
      name: "variable_scope",
      type: "string",
      schema: variableScopeSchema,
    },
    {
      name: "variable_operation",
      type: "string",
      schema: z
        .enum(["overwrite", "append", "prepend", "increment", "decrement"])
        .describe(
          "How the value is combined with the stored one. increment and decrement add or subtract a number."
        ),
    },
    {
      name: "variable_value",
      type: "string",
      schema: templated(z.string(), "Value to store."),
    },
  ],
  result: { schema: z.unknown().describe("The new value of the variable.") },
  run: { kind: "custom" },
};
