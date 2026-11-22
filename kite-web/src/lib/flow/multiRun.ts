import { Edge, Node } from "@xyflow/react";
import { getBlockDefinition } from "../blocks";
import { NodeData } from "./dataSchema";
import { normalizeHandle } from "./nodes";
import { isNoteNode } from "./notes";
import { walkDownstream } from "./placeholders";
import { isResumeEdge } from "./resume";

// Finds blocks that can run more than once in a single run because multiple
// paths from the same entry or button/modal branch converge into them, and
// returns the worst-case run count per block. Mutually exclusive merges
// (e.g. If yes/no, AI Check, Error Handler, single-match Conditions, or
// separate buttons) are not flagged.
export function getMultiRunCounts(
  flowNodes: Node<NodeData>[],
  edges: Edge[]
): Map<string, number> {
  const result = new Map<string, number>();
  const nodes = flowNodes.filter((n) => !isNoteNode(n));
  const byId = new Map(nodes.map((n) => [n.id, n]));

  const incomingByTarget = new Map<string, Edge[]>();
  const outSameRun = new Map<string, Edge[]>();
  const resumeGroups = new Map<string, { sourceId: string; edges: Edge[] }>();

  for (const edge of edges) {
    const source = byId.get(edge.source);
    const target = byId.get(edge.target);
    if (!source || !target) continue;

    const incoming = incomingByTarget.get(edge.target);
    if (incoming) {
      incoming.push(edge);
    } else {
      incomingByTarget.set(edge.target, [edge]);
    }

    if (isResumeEdge(edge, source.type)) {
      const handle = normalizeHandle(edge.sourceHandle) ?? "default";
      const key = `${edge.source}:${handle}`;
      const group = resumeGroups.get(key);
      if (group) {
        group.edges.push(edge);
      } else {
        resumeGroups.set(key, { sourceId: edge.source, edges: [edge] });
      }
    } else {
      const outgoing = outSameRun.get(edge.source);
      if (outgoing) {
        outgoing.push(edge);
      } else {
        outSameRun.set(edge.source, [edge]);
      }
    }
  }

  const candidates: [string, Edge[]][] = [];
  for (const [targetId, incoming] of incomingByTarget) {
    if (incoming.length < 2) continue;
    const type = byId.get(targetId)?.type ?? "";
    if (type.startsWith("entry_") || type.startsWith("option_")) continue;
    candidates.push([targetId, incoming]);
  }
  if (candidates.length === 0) return result;

  const entryIds = nodes
    .filter((n) => n.type?.startsWith("entry_"))
    .map((n) => n.id);
  if (entryIds.length === 0) return result;

  const reachable = new Set([...entryIds, ...walkDownstream(entryIds, edges)]);
  const reachableResumeGroups = [...resumeGroups.values()].filter((g) =>
    reachable.has(g.sourceId)
  );

  const countArrivals = (targetId: string, onlyEdge?: Edge) => {
    const memo = new Map<string, number>();
    const visiting = new Set<string>();

    const evalEdge = (edge: Edge): number => {
      if (edge.target === targetId) {
        return !onlyEdge || edge === onlyEdge ? 1 : 0;
      }
      return visit(edge.target);
    };

    const visit = (nodeId: string): number => {
      if (nodeId === targetId) return 0;
      const cached = memo.get(nodeId);
      if (cached !== undefined) return cached;
      if (visiting.has(nodeId)) return 0;

      visiting.add(nodeId);
      const node = byId.get(nodeId);
      const outEdges = outSameRun.get(nodeId) ?? [];
      const component = getBlockDefinition(node?.type)?.component;

      let total = 0;
      if (component === "condition") {
        const items: number[] = [];
        const elseBranches: number[] = [];
        for (const edge of outEdges) {
          const val = evalEdge(edge);
          if (byId.get(edge.target)?.type === "control_condition_item_else") {
            elseBranches.push(val);
          } else {
            items.push(val);
          }
        }
        const elseMax = elseBranches.length ? Math.max(...elseBranches) : 0;
        if (node?.data?.condition_allow_multiple) {
          total = Math.max(
            items.reduce((sum, v) => sum + v, 0),
            elseMax
          );
        } else {
          total = Math.max(0, ...items, elseMax);
        }
      } else if (component === "branches") {
        // Error handler is also a "branches" block: default and error can both
        // run in credits.ts, but merging them to continue after catching an
        // error is intentional rather than a duplicate path.
        const byHandle = new Map<string, number>();
        for (const edge of outEdges) {
          const handle = normalizeHandle(edge.sourceHandle) ?? "default";
          byHandle.set(handle, (byHandle.get(handle) ?? 0) + evalEdge(edge));
        }
        total = Math.max(0, ...byHandle.values());
      } else {
        for (const edge of outEdges) {
          total += evalEdge(edge);
        }
      }

      visiting.delete(nodeId);
      memo.set(nodeId, total);
      return total;
    };

    return {
      fromNode: visit,
      fromEdges: (startEdges: Edge[]) =>
        startEdges.reduce((sum, e) => sum + evalEdge(e), 0),
    };
  };

  for (const [targetId, incoming] of candidates) {
    const allCounter = countArrivals(targetId);
    const singleCounters = incoming.map((e) => countArrivals(targetId, e));

    let worst = 0;

    for (const entryId of entryIds) {
      const total = allCounter.fromNode(entryId);
      if (total <= 1) continue;
      const maxSingle = Math.max(
        0,
        ...singleCounters.map((c) => c.fromNode(entryId))
      );
      // Compare against maxSingle so a valid mutually-exclusive merge
      // downstream of an earlier diamond isn't blamed as a second convergence.
      if (total > maxSingle) {
        worst = Math.max(worst, total);
      }
    }

    for (const group of reachableResumeGroups) {
      const total = allCounter.fromEdges(group.edges);
      if (total <= 1) continue;
      const maxSingle = Math.max(
        0,
        ...singleCounters.map((c) => c.fromEdges(group.edges))
      );
      if (total > maxSingle) {
        worst = Math.max(worst, total);
      }
    }

    if (worst > 1) {
      result.set(targetId, worst);
    }
  }

  return result;
}

export function formatMultiRunWarning(count: number): string {
  return `More than one path connects into this block, so it and the blocks after it can run up to ${count} times per run. Remove an extra connection if it should only run once.`;
}

interface NodeSignature {
  id: string;
  type: string | undefined;
  allowMultiple: boolean;
}

let cachedNodes: Node<NodeData>[] | null = null;
let cachedEdges: Edge[] | null = null;
let cachedNodeSigs: NodeSignature[] = [];
let cachedResult = new Map<string, number>();

function matchesCachedSignatures(
  flowNodes: Node<NodeData>[],
  sigs: NodeSignature[]
) {
  if (flowNodes.length !== sigs.length) return false;
  for (let i = 0; i < flowNodes.length; i++) {
    const n = flowNodes[i];
    const sig = sigs[i];
    if (
      n.id !== sig.id ||
      n.type !== sig.type ||
      Boolean(n.data?.condition_allow_multiple) !== sig.allowMultiple
    ) {
      return false;
    }
  }
  return true;
}

// Cached wrapper so 60fps node drags (which only update positions) don't
// recompute the graph or re-render marker/editor subscribers.
export function getCachedMultiRunCounts(
  flowNodes: Node<NodeData>[],
  edges: Edge[]
): Map<string, number> {
  if (cachedNodes === flowNodes && cachedEdges === edges) {
    return cachedResult;
  }

  if (
    cachedEdges === edges &&
    matchesCachedSignatures(flowNodes, cachedNodeSigs)
  ) {
    cachedNodes = flowNodes;
    return cachedResult;
  }

  cachedNodes = flowNodes;
  cachedEdges = edges;
  cachedNodeSigs = flowNodes.map((n) => ({
    id: n.id,
    type: n.type,
    allowMultiple: Boolean(n.data?.condition_allow_multiple),
  }));
  cachedResult = getMultiRunCounts(flowNodes, edges);
  return cachedResult;
}
