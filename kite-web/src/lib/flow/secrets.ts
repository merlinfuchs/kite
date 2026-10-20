import { FlowData } from "../types/flow.gen";

// The names of the secrets a flow references, like API_KEY in
// {{secrets.API_KEY}}. Only requests and the values of comparison conditions
// can use secrets, and only placeholders reference them.
export function getReferencedSecretNames(flow: FlowData) {
  const names = new Set<string>();
  for (const node of flow.nodes) {
    let data: string;
    if (node.type === "action_http_request") {
      data = JSON.stringify(node.data);
    } else if (node.type === "control_condition_item_compare") {
      data = node.data.condition_item_value ?? "";
    } else {
      continue;
    }

    for (const [, placeholder] of data.matchAll(/\{\{(.*?)\}\}/g)) {
      for (const [, name] of placeholder.matchAll(
        /\bsecrets\.([A-Za-z_][A-Za-z0-9_]*)/g
      )) {
        names.add(name);
      }
    }
  }
  return Array.from(names);
}
