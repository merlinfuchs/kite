import { variableIdSchema, variableScopeSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableDelete: BlockDefinition = {
  type: "action_variable_delete",
  title: "Delete stored variable",
  description: "Delete the value of a stored variable",
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
  run: { kind: "custom" },
};
