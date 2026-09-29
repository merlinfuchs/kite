import { FlowData } from "../types/flow.gen";

// The names of the secrets a flow references, like API_KEY in
// {{secrets.API_KEY}}.
export function getReferencedSecretNames(flow: FlowData) {
  const names = new Set<string>();
  for (const m of JSON.stringify(flow.nodes).matchAll(
    /\bsecrets\.([A-Za-z0-9_]+)/g
  )) {
    names.add(m[1]);
  }
  return Array.from(names);
}
