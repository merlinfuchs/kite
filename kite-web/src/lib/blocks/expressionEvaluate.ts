import { nodeActionExpressionEvaluateDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const expressionEvaluate: BlockDefinition = {
  type: "action_expression_evaluate",
  title: "Calculate Value",
  description:
    "Evaluate math or other logical expressions and use the result later",
  icon: "calculator",
  category: "Utilities",
  credits: 1,
  schema: nodeActionExpressionEvaluateDataSchema,
  inputs: ["expression", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
