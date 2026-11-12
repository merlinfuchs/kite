import { FlowNodeData } from "../types/flow.gen";

// The IDs of the stored variables a block uses: its variable_id, and those of
// a Run JavaScript block's variables. Matches FlowNodeData.VariableIDs in the
// service.
export function nodeVariableIds(data: FlowNodeData | undefined): string[] {
  const ids = data?.variable_id ? [data.variable_id] : [];
  for (const v of data?.javascript_data?.variables ?? []) {
    if (v.variable_id) ids.push(v.variable_id);
  }
  return ids;
}

// data with each ID of nodeVariableIds replaced by what map returns for it,
// or cleared for undefined.
export function mapNodeVariableIds<T extends FlowNodeData>(
  data: T,
  map: (id: string) => string | undefined
): T {
  const res = { ...data };
  if (res.variable_id) {
    res.variable_id = map(res.variable_id);
    if (!res.variable_id) delete res.variable_id;
  }
  if (res.javascript_data?.variables) {
    res.javascript_data = {
      ...res.javascript_data,
      variables: res.javascript_data.variables.map((v) => ({
        ...v,
        variable_id: v.variable_id ? map(v.variable_id) : undefined,
      })),
    };
  }
  return res;
}
