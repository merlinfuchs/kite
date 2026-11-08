import { getBlockDefinition } from "../blocks";
import { NodeData } from "./dataSchema";
import { discordEventType } from "./eventTypes";
import { getNodeCreditsCost, getNodeValues, normalizeHandle } from "./nodes";

type FlowNode = { id: string; type?: string; data: NodeData };
type FlowEdge = {
  source: string;
  target: string;
  sourceHandle?: string | null;
};
type Cost = { min: number; max: number };

const free: Cost = { min: 0, max: 0 };

export interface FlowCredits {
  // The fewest and most credits one run can cost, taking one branch of each
  // condition. A loop counts as often as it runs when that's a number.
  min: number;
  max: number;
  // A loop runs a number of times only known when the flow runs, which max
  // counts as once.
  variableLoops: boolean;
}

// getFlowCredits estimates the credits of one run of a flow, the way the
// service counts them in CreditsCost. The branches of buttons run when
// they're clicked, so they aren't part of the run.
export function getFlowCredits(
  nodes: FlowNode[],
  edges: FlowEdge[]
): FlowCredits {
  const byId = new Map(nodes.map((n) => [n.id, n]));
  const children = new Map<string, FlowNode[]>();
  for (const edge of edges) {
    const target = byId.get(edge.target);
    if (!target) continue;
    const key = `${edge.source}:${normalizeHandle(edge.sourceHandle)}`;
    children.set(key, [...(children.get(key) ?? []), target]);
  }
  const childrenOf = (node: FlowNode, handle = "default") =>
    children.get(`${node.id}:${normalizeHandle(handle)}`) ?? [];
  const run = (node: FlowNode, handle?: string) =>
    sum(childrenOf(node, handle).map(cost));

  let variableLoops = false;

  const after = (node: FlowNode): Cost => {
    const component = getBlockDefinition(node.type)?.component;

    if (component === "condition") {
      // Without multiple matches only one branch runs, and none when no
      // branch matches and there is no else.
      // With multiple matches every matching branch runs, and else only
      // when none matched.
      const isElse = (i: FlowNode) => i.type === "control_condition_item_else";
      const items = childrenOf(node)
        .filter((i) => !isElse(i))
        .map(cost);
      const otherwise = childrenOf(node).filter(isElse).map(cost);
      const all = [...items, ...otherwise];
      return {
        min: otherwise.length ? Math.min(...all.map((c) => c.min)) : 0,
        max: node.data.condition_allow_multiple
          ? Math.max(
              items.reduce((s, c) => s + c.max, 0),
              ...otherwise.map((c) => c.max)
            )
          : Math.max(0, ...all.map((c) => c.max)),
      };
    }

    if (component === "control_loop") {
      const count = Number(node.data.loop_count);
      const known = Number.isInteger(count) && count >= 0;
      if (!known) variableLoops = true;
      const isEach = (c: FlowNode) =>
        getBlockDefinition(c.type)?.component === "control_loop_each";
      const each = sum(childrenOf(node).filter(isEach).map(cost));
      const end = sum(
        childrenOf(node)
          .filter((c) => !isEach(c))
          .map(cost)
      );
      return {
        min: end.min + (known ? each.min * count : 0),
        max: end.max + each.max * (known ? count : 1),
      };
    }

    // The only block that can run more than one of its outputs: the error
    // branch runs after some of the tried blocks failed, which were charged
    // too. The fewest credits are those of a run without errors.
    if (node.type === "control_error_handler") {
      const tried = run(node);
      const error = run(node, "error");
      return { min: tried.min, max: tried.max + error.max };
    }

    const outputs = getNodeValues(node.type ?? "").outputs ?? ["default"];
    if (!outputs.length) return free;
    const branches = outputs.map((h) => run(node, h));
    return {
      min: Math.min(...branches.map((b) => b.min)),
      max: Math.max(...branches.map((b) => b.max)),
    };
  };

  // A block reached on several paths is counted on each of them, but only
  // walked once. A flow that loops back to an earlier block counts it once.
  const known = new Map<string, Cost>();
  const visiting = new Set<string>();

  const cost = (node: FlowNode): Cost => {
    const cached = known.get(node.id);
    if (cached) return cached;
    if (visiting.has(node.id)) return free;

    visiting.add(node.id);
    const rest = after(node);
    visiting.delete(node.id);

    const own = nodeCredits(node);
    const total = { min: own + rest.min, max: own + rest.max };
    known.set(node.id, total);
    return total;
  };

  const entry = nodes.find((n) => n.type?.startsWith("entry_"));
  return { ...(entry ? cost(entry) : free), variableLoops };
}

function nodeCredits(node: FlowNode) {
  const cost = getNodeCreditsCost(getNodeValues(node.type ?? ""), node.data);
  if (cost !== undefined) return cost;
  // The service charges 1 for actions without a cost.
  return node.type?.startsWith("action_") ? 1 : 0;
}

function sum(costs: Cost[]) {
  return costs.reduce(
    (s, c) => ({ min: s.min + c.min, max: s.max + c.max }),
    free
  );
}

// chargedFrequentEvent is what a listener for a frequent event is called,
// like "message", if it costs credits even for the events it has nothing to
// do for, because no event filter or condition comes before its first paid
// block.
export function chargedFrequentEvent(
  nodes: FlowNode[],
  edges: FlowEdge[],
  credits: FlowCredits
): string | undefined {
  const entry = nodes.find((n) => n.type === "entry_event");
  const frequent = discordEventType(entry?.data.event_type)?.frequent;
  if (!entry || !frequent || credits.min === 0) return undefined;

  const filtered = edges.some(
    (e) =>
      e.target === entry.id &&
      nodes.find((n) => n.id === e.source)?.type === "option_event_filter"
  );
  return filtered ? undefined : frequent;
}

// formatFlowCredits is the most credits a run costs, like "up to 3 credits".
export function formatFlowCredits(credits: FlowCredits) {
  if (credits.max === 0) return "free";
  const n = credits.max === 1 ? "1 credit" : `${credits.max} credits`;
  return credits.min === credits.max ? n : `up to ${n}`;
}
