import { z } from "zod";
import {
  numericOrPlaceholder,
  templated,
  variableIdSchema,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const variableListSorts = [
  { value: "value_desc", label: "Highest value first" },
  { value: "value_asc", label: "Lowest value first" },
  { value: "updated_desc", label: "Last updated first" },
  { value: "updated_asc", label: "First updated first" },
] as const;

export const variableList: BlockDefinition = {
  type: "action_variable_list",
  id_prefix: "values",
  title: "List stored variable values",
  description:
    "Get the values of a stored variable in order, like a leaderboard of the highest ones",
  icon: "list-ordered",
  category: "Stored variables",
  // Unlike the other reads, it sorts every value of the variable.
  credits: 1,
  strict_settings: true,
  fields: [
    { name: "variable_id", schema: variableIdSchema },
    {
      name: "variable_sort",
      input: "variable_list",
      schema: z
        .enum(["value_desc", "value_asc", "updated_desc", "updated_asc"])
        .optional()
        .describe(
          "Order of the values. value_desc and value_asc sort numbers, and text of a number like 42, from highest or lowest, with other values last. updated_desc and updated_asc sort by when each value last changed. Defaults to value_desc."
        ),
    },
    {
      name: "variable_scope_prefix",
      input: "variable_list",
      schema: templated(
        z.string().optional(),
        "Only values whose scope starts with this, like {{guild.id}}: for the members of one server when the scope is {{guild.id}}:{{user.id}}. All values if empty, and ignored for variables without a scope."
      ),
    },
    {
      name: "variable_limit",
      input: "variable_list",
      schema: numericOrPlaceholder(
        "How many values to get, from 1 to 100. Defaults to 10.",
        /^(100|[1-9][0-9]?)$/
      ).optional(),
    },
    {
      name: "variable_skip",
      input: "variable_list",
      schema: numericOrPlaceholder(
        "How many values to skip first, from 0 to 10000, like 10 for the second page of 10. Defaults to 0.",
        /^(10000|[1-9]?[0-9]{1,3})$/
      ).optional(),
    },
  ],
  result: {
    schema: z
      .array(
        z.object({
          scope: z
            .string()
            .nullable()
            .describe("The scope of the value, null for unscoped variables."),
          value: z.unknown().describe("The stored value."),
          updated_at: z
            .number()
            .describe(
              "When the value last changed, as a Unix timestamp in seconds."
            ),
        })
      )
      .describe("The values in the chosen order."),
  },
  run: { kind: "custom" },
};
