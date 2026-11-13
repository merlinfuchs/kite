import { getBlockDefinition } from "../blocks";
import { BlockField } from "../blocks/types";
import { FlowData } from "../types/flow.gen";
import { templateSecretNames } from "./templateSecrets";

// The setting, or the part of it like javascript_data.code, that can use
// secrets, see BlockField.secrets.
export function secretsPath(field: BlockField) {
  if (!field.secrets) return undefined;
  return typeof field.secrets === "string"
    ? `${field.name}.${field.secrets}`
    : field.name;
}

// The names of the secrets a flow references, like API_KEY in
// {{secrets.API_KEY}}. Only the settings of blocks that allow secrets can use
// them, and only placeholders reference them.
export function getReferencedSecretNames(flow: FlowData) {
  const names = new Set<string>();
  for (const node of flow.nodes) {
    const fields = getBlockDefinition(node.type)?.fields ?? [];
    for (const field of fields) {
      const path = secretsPath(field);
      if (!path) continue;
      const value = path
        .split(".")
        .reduce<unknown>(
          (v, key) => (v as Record<string, unknown> | undefined)?.[key],
          node.data
        );
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
