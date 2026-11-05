import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableColumnDefinitionFields } from "./tableFields";

export const tableColumnAdd: BlockDefinition = {
  type: "action_table_column_add",
  id_prefix: "tca",
  title: "Add table column",
  description: "Add a column to a table.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: tableColumnDefinitionFields,
  result: { schema: z.unknown().describe("The added table column.") },
  run: { kind: "custom" },
  strict_settings: true,
};
