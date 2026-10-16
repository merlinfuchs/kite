import { FlowData, FunctionParameterData } from "../types/flow.gen";
import {
  Command,
  EventListener,
  FlowFunction,
  Message,
} from "../types/wire.gen";

// The parameters of a function, defined by the entry block of its flow. The
// Call function block passes a value for each of them.
export function getFunctionParameters(
  flow: FlowData | undefined
): FunctionParameterData[] {
  const entry = flow?.nodes?.find((n) => n.type === "entry_function");
  return (entry?.data?.function_parameters ?? []).filter((p) => p.name);
}

// A function as it's stored in share codes and exported JSON. The id is the
// function's id in the app it was exported from, so flows in the same share
// can be pointed at the imported copy.
export interface SharedFunction {
  id: string;
  flow_source: FlowData;
}

export function functionToShared(fn: FlowFunction): SharedFunction {
  return { id: fn.id, flow_source: fn.flow_source };
}

export function isSharedFunction(value: unknown): value is SharedFunction {
  const f = value as SharedFunction | null | undefined;
  return (
    !!f &&
    typeof f === "object" &&
    typeof f.id === "string" &&
    !!f.flow_source &&
    Array.isArray(f.flow_source.nodes) &&
    Array.isArray(f.flow_source.edges) &&
    f.flow_source.nodes.some((n) => n?.type === "entry_function")
  );
}

// The name of a function, set on the entry block of its flow.
export function getFunctionName(flow: FlowData | null | undefined) {
  const entry = flow?.nodes?.find((n) => n.type === "entry_function");
  return entry?.data?.name || "Unnamed function";
}

export function getReferencedFunctionIds(flow: FlowData | null | undefined) {
  const ids = new Set<string>();
  for (const node of flow?.nodes ?? []) {
    if (node.data?.function_id) {
      ids.add(node.data.function_id);
    }
  }
  return ids;
}

// The bundled functions an import needs because the app doesn't have them:
// those the flows call, and those called by those functions in turn.
// Functions that aren't bundled can't be imported and are left out.
export function getMissingFunctions(
  flows: (FlowData | null | undefined)[],
  bundled: SharedFunction[],
  existingIds: Set<string>
): SharedFunction[] {
  const byId = new Map(bundled.map((f) => [f.id, f]));
  const seen = new Set<string>();
  const res: SharedFunction[] = [];

  const queue = flows.flatMap((f) => Array.from(getReferencedFunctionIds(f)));
  while (queue.length > 0) {
    const id = queue.shift()!;
    if (seen.has(id) || existingIds.has(id)) continue;
    seen.add(id);

    const fn = byId.get(id);
    if (!fn) continue;

    res.push(fn);
    queue.push(...Array.from(getReferencedFunctionIds(fn.flow_source)));
  }

  return res;
}

// Something that calls a function, shown before the function is deleted.
export interface FunctionUsage {
  kind: "command" | "event_listener" | "message" | "function";
  id: string;
  label: string;
}

function callsFunction(flow: FlowData | null | undefined, functionId: string) {
  return getReferencedFunctionIds(flow).has(functionId);
}

// The commands, event listeners, message templates and other functions that
// call the function. A function calling itself doesn't count.
export function getFunctionUsages(
  functionId: string,
  {
    commands,
    eventListeners,
    messages,
    functions,
  }: {
    commands: (Command | undefined)[];
    eventListeners: (EventListener | undefined)[];
    messages: (Message | undefined)[];
    functions: (FlowFunction | undefined)[];
  }
): FunctionUsage[] {
  const res: FunctionUsage[] = [];

  for (const c of commands) {
    if (c && callsFunction(c.flow_source, functionId)) {
      res.push({ kind: "command", id: c.id, label: `/${c.name}` });
    }
  }
  for (const l of eventListeners) {
    if (l && callsFunction(l.flow_source, functionId)) {
      res.push({
        kind: "event_listener",
        id: l.id,
        label: `event listener '${l.description || l.type}'`,
      });
    }
  }
  for (const m of messages) {
    if (
      m &&
      Object.values(m.flow_sources ?? {}).some((f) =>
        callsFunction(f, functionId)
      )
    ) {
      res.push({
        kind: "message",
        id: m.id,
        label: `message template '${m.name}'`,
      });
    }
  }
  for (const f of functions) {
    if (f && f.id !== functionId && callsFunction(f.flow_source, functionId)) {
      res.push({ kind: "function", id: f.id, label: `function '${f.name}'` });
    }
  }

  return res;
}

// Lists usages in a sentence, like "/ban, /kick and 2 more".
export function describeFunctionUsages(usages: FunctionUsage[], max = 5) {
  const labels = usages.slice(0, max).map((u) => u.label);
  const more = usages.length - labels.length;
  if (more > 0) labels.push(`${more} more`);
  if (labels.length <= 1) return labels.join("");
  return `${labels.slice(0, -1).join(", ")} and ${labels[labels.length - 1]}`;
}
