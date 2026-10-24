import { getBlockDefinition } from "../blocks";
import { NodeData } from "./dataSchema";
import { getNodeCreditsCost, getNodeValues } from "./nodes";
import { Template } from "./templates";

type FlowNode = { id: string; type?: string; data: NodeData };
type FlowEdge = {
  source: string;
  target: string;
  sourceHandle?: string | null;
};

export interface FlowCredits {
  // The fewest and most credits one run can cost, taking one branch of each
  // condition. A loop counts as often as it runs when that's a number.
  min: number;
  max: number;
  // A loop runs a number of times only known when the flow runs, which max
  // counts as once.
  variableLoops: boolean;
}

// Events that can happen many times a minute in a busy server, where a flow
// that costs credits on every run adds up quickly, and what one is called.
export const frequentEvents: Record<string, string> = {
  message_create: "message",
  message_update: "edited message",
  message_delete: "deleted message",
  message_reaction_add: "reaction",
  message_reaction_remove: "removed reaction",
};

// getFlowCredits estimates the credits of one run of a flow, the way the
// service counts them in CreditsCost. The branches of buttons run when
// they're clicked, so they aren't part of the run.
export function getFlowCredits(
  nodes: FlowNode[],
  edges: FlowEdge[]
): FlowCredits {
  const byId = new Map(nodes.map((n) => [n.id, n]));
  const children = new Map<string, Map<string, FlowNode[]>>();
  for (const edge of edges) {
    const target = byId.get(edge.target);
    if (!target) continue;
    const handle = edge.sourceHandle || "default";
    if (!children.has(edge.source)) children.set(edge.source, new Map());
    const handles = children.get(edge.source)!;
    if (!handles.has(handle)) handles.set(handle, []);
    handles.get(handle)!.push(target);
  }

  let variableLoops = false;
  const visiting = new Set<string>();

  const cost = (node: FlowNode): { min: number; max: number } => {
    // A flow that loops back to an earlier block is counted once.
    if (visiting.has(node.id)) return { min: 0, max: 0 };
    visiting.add(node.id);

    const handles = children.get(node.id) ?? new Map<string, FlowNode[]>();
    const run = (handle: string) => sum((handles.get(handle) ?? []).map(cost));
    const own = nodeCredits(node);
    const block = getBlockDefinition(node.type);

    let after: { min: number; max: number };
    if (block?.component === "condition") {
      // Without multiple matches only one branch runs, and none when no
      // branch matches and there is no else.
      const items = (handles.get("default") ?? []).map((item) => ({
        item,
        ...cost(item),
      }));
      const hasElse = items.some(
        (i) => i.item.type === "control_condition_item_else"
      );
      after = {
        min: hasElse ? Math.min(...items.map((i) => i.min)) : 0,
        max: node.data.condition_allow_multiple
          ? items.reduce((s, i) => s + i.max, 0)
          : Math.max(0, ...items.map((i) => i.max)),
      };
    } else if (node.type === "control_loop") {
      const count = Number(node.data.loop_count);
      const times = Number.isInteger(count) && count >= 0 ? count : 1;
      if (times !== count) variableLoops = true;
      const each = sum(
        (handles.get("default") ?? [])
          .filter((c) => c.type === "control_loop_each")
          .map(cost)
      );
      const end = sum(
        (handles.get("default") ?? [])
          .filter((c) => c.type !== "control_loop_each")
          .map(cost)
      );
      after = {
        min: end.min + (times === count ? each.min * times : 0),
        max: end.max + each.max * times,
      };
    } else if (node.type === "control_error_handler") {
      // The error branch runs after some of the others failed.
      const tried = run("default");
      const error = run("error");
      after = {
        min: Math.min(tried.min, error.min),
        max: tried.max + error.max,
      };
    } else {
      const outputs = getNodeValues(node.type ?? "").outputs ?? ["default"];
      const branches = outputs.map(run);
      after =
        outputs.length > 1
          ? {
              min: Math.min(...branches.map((b) => b.min)),
              max: Math.max(...branches.map((b) => b.max)),
            }
          : branches[0] ?? { min: 0, max: 0 };
    }

    visiting.delete(node.id);
    return { min: own + after.min, max: own + after.max };
  };

  const entry = nodes.find((n) => n.type?.startsWith("entry_"));
  const total = entry ? cost(entry) : { min: 0, max: 0 };
  return { ...total, variableLoops };
}

function nodeCredits(node: FlowNode) {
  const cost = getNodeCreditsCost(getNodeValues(node.type ?? ""), node.data);
  if (cost !== undefined) return cost;
  // The service charges 1 for actions without a cost.
  return node.type?.startsWith("action_") ? 1 : 0;
}

function sum(costs: { min: number; max: number }[]) {
  return costs.reduce((s, c) => ({ min: s.min + c.min, max: s.max + c.max }), {
    min: 0,
    max: 0,
  });
}

// costsOnEveryFrequentEvent is whether a listener for a frequent event costs
// credits even for the events it has nothing to do for, because no event
// filter or condition comes before its first paid block.
export function costsOnEveryFrequentEvent(
  nodes: FlowNode[],
  edges: FlowEdge[],
  credits: FlowCredits
) {
  const entry = nodes.find((n) => n.type === "entry_event");
  if (!entry || !Object.hasOwn(frequentEvents, entry.data.event_type ?? "")) {
    return false;
  }
  const filtered = edges.some(
    (e) =>
      e.target === entry.id &&
      nodes.find((n) => n.id === e.source)?.type === "option_event_filter"
  );
  return !filtered && credits.min > 0;
}

// formatFlowCredits is the most credits a run costs, like "up to 3 credits".
export function formatFlowCredits(credits: FlowCredits) {
  if (credits.max === 0) return "free";
  const n = credits.max === 1 ? "1 credit" : `${credits.max} credits`;
  return credits.min === credits.max ? n : `up to ${n}`;
}

export type TemplateFlowCredits = {
  key: string;
  credits: FlowCredits;
} & (
  | { kind: "command"; name: string }
  | { kind: "event"; eventType: string }
  | { kind: "button"; messageKey: string; flowSourceId: string }
);

// getTemplateCredits estimates a run of each flow of a template, with what
// runs it.
export function getTemplateCredits(template: Template): TemplateFlowCredits[] {
  const estimate = (flow: { nodes: FlowNode[]; edges: FlowEdge[] }) =>
    getFlowCredits(flow.nodes, flow.edges);

  return [
    ...template.commands.map((c) => ({
      key: c.key,
      kind: "command" as const,
      name: c.name,
      credits: estimate(c.flowSource({})),
    })),
    ...template.eventListeners.map((l) => ({
      key: l.key,
      kind: "event" as const,
      eventType: l.type,
      credits: estimate(l.flowSource({})),
    })),
    ...template.messages.flatMap((m) =>
      Object.entries(m.flowSources({})).map(([id, flow]) => ({
        key: `${m.key}:${id}`,
        kind: "button" as const,
        messageKey: m.key,
        flowSourceId: id,
        credits: estimate(flow),
      }))
    ),
  ];
}
