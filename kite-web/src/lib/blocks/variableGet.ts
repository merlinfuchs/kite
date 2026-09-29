import { nodeActionVariableGetSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableGet: BlockDefinition = {
  type: "action_variable_get",
  title: "Get stored variable",
  description: "Get the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  schema: nodeActionVariableGetSchema,
  inputs: ["variable_id", "variable_scope", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
