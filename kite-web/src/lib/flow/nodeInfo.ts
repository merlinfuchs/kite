import { JsonSchema7Type } from "zod-to-json-schema";
import { resultJsonSchema, toJsonSchema } from "./catalog";
import { inlineTypeJsonSchema, JsonSchema } from "./expressionTypes";
import { getNodeCreditsCost, getNodeValues } from "./nodes";

export type NodeInfo = {
  title: string;
  description: string;
  color: string;
  icon: string;
  dataSchema: JsonSchema7Type | null;
  resultSchema: JsonSchema | null;
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
    resultSchema: values.result
      ? resultJsonSchema(values.result, inlineTypeJsonSchema)
      : null,
    dataFields: values.dataFields,
    creditsCost: getNodeCreditsCost(values, {}) ?? null,
  };
}
