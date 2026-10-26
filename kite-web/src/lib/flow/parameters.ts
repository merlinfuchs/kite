import { FlowData, FlowParameter } from "../types/flow.gen";
import { placeholderReferences } from "./templateSecrets";
import { isNoteNode } from "./notes";

// parameterKeyPattern is what the key of a parameter can be.
export const parameterKeyPattern = /^[A-Za-z_][A-Za-z0-9_]*$/;

// parameterKeys are the keys of the parameters a value uses as
// {{params.<key>}}, in the order they're used.
export function parameterKeys(value: string): string[] {
  return placeholderReferences(value, "params");
}

// Moving a block keeps its data, so only changed blocks are searched again.
const nodeKeys = new WeakMap<object, string[]>();

// referencedParameters are the keys of the parameters the blocks of the flow
// use. The service finds them the same way. Notes only mention them.
export function referencedParameters(flow: {
  nodes: { type?: string; data?: unknown }[];
}): string[] {
  const keys: string[] = [];
  for (const node of flow.nodes) {
    if (isNoteNode(node)) continue;
    const data = node.data && typeof node.data === "object" ? node.data : null;
    let own = data ? nodeKeys.get(data) : undefined;
    if (!own) {
      own = parameterKeys(JSON.stringify(node.data ?? {}));
      if (data) nodeKeys.set(data, own);
    }
    for (const key of own) {
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
  return withParameterKeys(parameters, referencedParameters(flow));
}

// withParameterKeys adds a parameter for every key there's none for.
export function withParameterKeys(
  parameters: FlowParameter[],
  keys: string[]
): FlowParameter[] {
  const missing = keys.filter((key) => !parameters.some((p) => p.key === key));
  if (!missing.length) return parameters;
  return [...parameters, ...missing.map(newParameter)];
}

// newParameter is a parameter the user just added, by using it in a block or
// in the sidebar. They add it because a block needs it, so it's required.
export function newParameter(key: string): FlowParameter {
  return { key, type: "text", required: true, value: "" };
}

// needsValue is whether the parameter is required but has no value.
export function needsValue(parameter: FlowParameter): boolean {
  return !!parameter.required && !parameter.value.trim();
}

// withoutParameterValues is the flow without the values of its parameters,
// as it's exported. They belong to the app it comes from.
export function withoutParameterValues<
  T extends { parameters?: FlowParameter[] }
>(flow: T): T {
  if (!flow.parameters?.length) return flow;
  return {
    ...flow,
    parameters: flow.parameters.map((p) => ({ ...p, value: "" })),
  };
}
