import { nodeActionVariableSetSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableSet: BlockDefinition = {
  type: "action_variable_set",
  title: "Set stored variable",
  description: "Set the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  schema: nodeActionVariableSetSchema,
  inputs: [
    "variable_id",
    "variable_scope",
    "variable_operation",
    "variable_value",
    "temporary_name",
    "custom_label",
  ],
  run: { kind: "custom" },
};
