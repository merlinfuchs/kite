import { FlowData, FlowParameter } from "../types/flow.gen";

const templatePattern = /{{(.*?)}}/g;
// Not after a dot, as webhook.data.params.count isn't a parameter.
const referencePattern = /(?:^|[^.\w])params\.([A-Za-z_][A-Za-z0-9_]*)/g;

// parameterKeyPattern is what the key of a parameter can be.
export const parameterKeyPattern = /^[A-Za-z_][A-Za-z0-9_]*$/;

// parameterKeys are the keys of the parameters a value uses as
// {{params.<key>}}, in the order they're used.
export function parameterKeys(value: string): string[] {
  const keys: string[] = [];
  for (const [, template] of value.matchAll(templatePattern)) {
    for (const [, key] of template.matchAll(referencePattern)) {
      if (!keys.includes(key)) keys.push(key);
    }
  }
  return keys;
}

// referencedParameters are the keys of the parameters the blocks of the flow
// use. The service finds them the same way.
export function referencedParameters(flow: {
  nodes: { data?: unknown }[];
}): string[] {
  const keys: string[] = [];
  for (const node of flow.nodes) {
    for (const key of parameterKeys(JSON.stringify(node.data ?? {}))) {
      if (!keys.includes(key)) keys.push(key);
    }
  }
  return keys;
}

// syncParameters returns the parameters with one for every key the flow uses,
// keeping the ones it has. Ones the flow doesn't use anymore stay, so a value
// isn't lost while a block is being changed.
export function syncParameters(
  flow: Pick<FlowData, "nodes">,
  parameters: FlowParameter[]
): FlowParameter[] {
  const missing = referencedParameters(flow).filter(
    (key) => !parameters.some((p) => p.key === key)
  );
  if (!missing.length) return parameters;
  return [
    ...parameters,
    ...missing.map((key) => ({ key, type: "text" as const, value: "" })),
  ];
}

// withoutParameterValues leaves out the values of the parameters of every flow
// in the data, like an export. They belong to the app it comes from.
export function withoutParameterValues<T>(data: T): T {
  if (Array.isArray(data)) {
    return data.map((v) => withoutParameterValues(v)) as T;
  }
  if (!data || typeof data !== "object") return data;

  const res: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(data)) {
    res[key] =
      key === "parameters" && Array.isArray(value) && "nodes" in data
        ? value.map((p) => ({ ...p, value: "" }))
        : withoutParameterValues(value);
  }
  return res as T;
}
