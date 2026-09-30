import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const expressionEvaluate: BlockDefinition = {
  type: "action_expression_evaluate",
  title: "Calculate Value",
  description:
    "Evaluate math or other logical expressions and use the result later",
  icon: "calculator",
  category: "Utilities",
  credits: 1,
  fields: [
    {
      name: "expression",
      schema: templated(
        z
          .string()
          .max(2000)
          .refine((val) => !val.startsWith("{{"), {
            message:
              "In most cases, you don't need to use the double curly brackets around the expression here. Only use them if you want to include a placeholder in the expression.",
          }),
        "Expr language expression to evaluate, written without surrounding {{ }}, e.g. arg('a') + arg('b')."
      ),
    },
  ],
  result: {
    schema: z.unknown().describe("The value the expression evaluates to."),
  },
  run: { kind: "custom" },
};
