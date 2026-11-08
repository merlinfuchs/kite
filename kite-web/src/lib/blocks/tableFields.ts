import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockField } from "./types";

export const tableFields: BlockField[] = [
  { name: "table_id", schema: templated(z.string().min(1), "The table ID.") },
];

export const tableRowFields: BlockField[] = [
  ...tableFields,
  { name: "table_row_id", schema: templated(z.string().min(1), "The row ID.") },
];

export const tableDataField: BlockField = {
  name: "table_data",
  schema: templated(z.string(), "A JSON object containing row column values."),
};

export const tableColumnIdField: BlockField = {
  name: "table_column_id",
  schema: templated(z.string().min(1), "The column ID."),
};

export const tableColumnFields: BlockField[] = [
  ...tableFields,
  tableColumnIdField,
];

export const tableColumnDefinitionFields: BlockField[] = [
  ...tableFields,
  {
    name: "table_column_name",
    schema: templated(z.string().min(1), "The column name."),
  },
  {
    name: "table_column_type",
    schema: templated(z.string().min(1), "The column type."),
  },
];
