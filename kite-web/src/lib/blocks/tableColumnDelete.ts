import { BlockDefinition } from "./types";
import { tableColumnFields } from "./tableFields";

export const tableColumnDelete: BlockDefinition = {
  type: "action_table_column_delete",
  id_prefix: "tcd",
  title: "Delete table column",
  description: "Delete a column from a table.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: tableColumnFields,
  run: { kind: "custom" },
  strict_settings: true,
};
