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
      let value = node.data[field.name];
      if (typeof field.secrets === "string") {
        value = (value as Record<string, unknown> | undefined)?.[field.secrets];
      }
      const data = JSON.stringify(value ?? "");
      for (const name of templateSecretNames(data)) {
        names.add(name);
      }
    }
  }
  return Array.from(names);
}

// missingSecretNames are the secrets the flows reference that aren't among
// the names.
export function missingSecretNames(
  flows: FlowData[],
  names: (string | undefined)[]
) {
  const have = new Set(names);
  return Array.from(
    new Set(flows.flatMap((f) => getReferencedSecretNames(f)))
  ).filter((name) => !have.has(name));
}
