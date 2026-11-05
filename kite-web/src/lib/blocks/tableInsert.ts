import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableDataField, tableFields } from "./tableFields";

export const tableInsert: BlockDefinition = {
  type: "action_table_insert",
  id_prefix: "tins",
  title: "Insert table row",
  description: "Create a row in a table.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: [...tableFields, tableDataField],
  result: { schema: z.unknown().describe("The created table row.") },
  run: { kind: "custom" },
  strict_settings: true,
};
