import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableDataField, tableRowFields } from "./tableFields";

export const tableUpdate: BlockDefinition = {
  type: "action_table_update",
  id_prefix: "tupd",
  title: "Update table row",
  description: "Replace the data in a table row.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: [...tableRowFields, tableDataField],
  result: { schema: z.unknown().describe("The updated table row.") },
  run: { kind: "custom" },
  strict_settings: true,
};
