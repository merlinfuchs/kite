import { getBlockDefinition } from "../blocks";
import { FlowData } from "../types/flow.gen";
import { templateSecretNames } from "./templateSecrets";

// The names of the secrets a flow references, like API_KEY in
// {{secrets.API_KEY}}. Only the settings of blocks that allow secrets can use
// them, and only placeholders reference them.
export function getReferencedSecretNames(flow: FlowData) {
  const names = new Set<string>();
  for (const node of flow.nodes) {
    const fields = getBlockDefinition(node.type)?.fields ?? [];
    for (const field of fields.filter((f) => f.secrets)) {
      const data = JSON.stringify(node.data[field.name] ?? "");
      for (const name of templateSecretNames(data)) {
        names.add(name);
      }
    }
  }
  return Array.from(names);
}
