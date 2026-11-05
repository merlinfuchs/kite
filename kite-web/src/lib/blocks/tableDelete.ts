import { BlockDefinition } from "./types";
import { tableRowFields } from "./tableFields";

export const tableDelete: BlockDefinition = {
  type: "action_table_delete",
  id_prefix: "tdel",
  title: "Delete table row",
  description: "Delete a row from a table.",
  icon: "database",
  category: "Tables",
  credits: 1,
  fields: tableRowFields,
  run: { kind: "custom" },
  strict_settings: true,
};
