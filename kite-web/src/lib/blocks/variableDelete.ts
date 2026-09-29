import { nodeActionVariableDeleteSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableDelete: BlockDefinition = {
  type: "action_variable_delete",
  title: "Delete stored variable",
  description: "Delete the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  schema: nodeActionVariableDeleteSchema,
  inputs: ["variable_id", "variable_scope", "custom_label"],
  run: { kind: "custom" },
};
