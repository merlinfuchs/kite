import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableFields } from "./tableFields";

export const tableColumns: BlockDefinition = {
  type: "action_table_columns",
  id_prefix: "tcols",
  title: "List table columns",
  description: "List the columns configured on a table.",
  icon: "database",
  category: "Tables",
  credits: 0,
  fields: tableFields,
  result: { schema: z.unknown().describe("The table columns.") },
  run: { kind: "custom" },
  strict_settings: true,
};
