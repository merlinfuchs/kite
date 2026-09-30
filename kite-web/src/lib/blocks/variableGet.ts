import { z } from "zod";
import { variableIdSchema, variableScopeSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableGet: BlockDefinition = {
  type: "action_variable_get",
  title: "Get stored variable",
  description: "Get the value of a stored variable",
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
  ],
  result: { schema: z.unknown().describe("The value of the variable.") },
  run: { kind: "custom" },
};
