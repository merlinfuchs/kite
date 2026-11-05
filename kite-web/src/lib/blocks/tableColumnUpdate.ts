import { z } from "zod";
import { BlockDefinition } from "./types";
import {
  tableColumnDefinitionFields,
  tableColumnIdField,
  tableFields,
} from "./tableFields";

export const tableColumnUpdate: BlockDefinition = {
  type: "action_table_column_update",
  id_prefix: "tcu",
  title: "Update table column",
  description: "Update a table column's name or type.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: [
    ...tableFields,
    tableColumnIdField,
    ...tableColumnDefinitionFields.slice(1),
  ],
  result: { schema: z.unknown().describe("The updated table column.") },
  run: { kind: "custom" },
  strict_settings: true,
};
