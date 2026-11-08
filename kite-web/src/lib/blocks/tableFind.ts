import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableFields } from "./tableFields";

export const tableFind: BlockDefinition = {
  type: "action_table_find",
  id_prefix: "tfind",
  title: "Find table rows",
  description: "Find rows in a table using JSON filters.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: [
    ...tableFields,
    {
      name: "table_filters",
      schema: z
        .string()
        .describe("A JSON array of filters with column, operator, and value."),
    },
    {
      name: "table_limit",
      schema: z.string().describe("Maximum rows to return, from 1 to 100."),
    },
  ],
  result: { schema: z.unknown().describe("The matching table rows.") },
  run: { kind: "custom" },
  strict_settings: true,
};
