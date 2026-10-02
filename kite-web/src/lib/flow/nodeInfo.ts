import { JsonSchema7Type } from "zod-to-json-schema";
import { toJsonSchema } from "./catalog";
import { getNodeCreditsCost, getNodeValues } from "./nodes";

export type NodeInfo = {
  title: string;
  description: string;
  color: string;
  icon: string;
  dataSchema: JsonSchema7Type | null;
  resultSchema: JsonSchema7Type | null;
  dataFields: string[];
  creditsCost: number | null;
};

// Info about a block shown by the docs.
export function getNodeInfo(nodeType: string): NodeInfo {
  const values = getNodeValues(nodeType);

  return {
    title: values.defaultTitle,
    description: values.defaultDescription,
    color: values.color,
    icon: values.icon,
    dataSchema: values.dataSchema ? toJsonSchema(values.dataSchema) : null,
    resultSchema: values.resultSchema
      ? toJsonSchema(values.resultSchema)
      : null,
    dataFields: values.dataFields,
    creditsCost: getNodeCreditsCost(values, {}) ?? null,
  };
}
