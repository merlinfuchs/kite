import { z } from "zod";
import { BlockDefinition } from "./types";
import { tableRowFields } from "./tableFields";

export const tableGet: BlockDefinition = {
  type: "action_table_get",
  id_prefix: "tget",
  title: "Get table row",
  description: "Get a row from a table by ID.",
  icon: "database",
  category: "Tables",
  credits: 0,
  fields: tableRowFields,
  result: {
    schema: z
      .unknown()
      .describe("The table row, or null when it does not exist."),
  },
  run: { kind: "custom" },
  strict_settings: true,
};
