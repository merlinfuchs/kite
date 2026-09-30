import { variableIdSchema, variableScopeSchema } from "../flow/dataSchema";
import { variableSettings } from "./fields";
import { BlockDefinition } from "./types";

export const variableDelete: BlockDefinition = {
  type: "action_variable_delete",
  title: "Delete stored variable",
  description: "Delete the value of a stored variable",
  icon: "variable",
  category: "Stored Variables",
  credits: 1,
  fields: [...variableSettings],
  run: { kind: "custom" },
};
