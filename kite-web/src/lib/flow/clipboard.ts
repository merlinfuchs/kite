import { Edge, Node, XYPosition } from "@xyflow/react";
import { NodeData } from "./dataSchema";
import { getEdgeId, getNodeId, getNodeValues } from "./nodes";

const clipboardType = "kite-flow-clipboard";

export interface FlowClipboard {
  type: typeof clipboardType;
  nodes: Node<NodeData>[];
  edges: Edge[];
}

// Collects the selected nodes plus the children owned by them (e.g. the
// condition items of a condition). Fixed nodes can only be copied along
// with their owner, so entry blocks are never copied.
export function copyFlowNodes(
  nodes: Node<NodeData>[],
  edges: Edge[]
): FlowClipboard | null {
  const copiedIds = withOwnedChildren(
    nodes.filter((n) => n.selected && !getNodeValues(n.type!).fixed),
    nodes,
    edges
  );

  if (copiedIds.size === 0) return null;

  return {
    type: clipboardType,
    nodes: nodes
      .filter((n) => copiedIds.has(n.id))
      .map(({ id, type, position, data }) => ({ id, type, position, data })),
    edges: edges
      .filter((e) => copiedIds.has(e.source) && copiedIds.has(e.target))
      .map(({ id, type, source, target, sourceHandle, targetHandle }) => ({
        id,
        type,
        source,
        target,
        sourceHandle,
        targetHandle,
      })),
  };
}

export function parseFlowClipboard(text: string): FlowClipboard | null {
  try {
    const value = JSON.parse(text);
    if (
      value?.type === clipboardType &&
      Array.isArray(value.nodes) &&
      Array.isArray(value.edges)
    ) {
      return value;
    }
  } catch {}
  return null;
}

// Creates fresh copies of the clipboard nodes with new IDs, moved so that
// their top left corner is at the given position (or slightly offset from the
// originals if none is given). Nodes rejected by
// isAllowed are dropped together with their edges and owned children.
export function pasteFlowNodes(
  clipboard: FlowClipboard,
  position: XYPosition | null,
  isAllowed: (type: string) => boolean = () => true
): [Node<NodeData>[], Edge[]] {
  const droppedIds = withOwnedChildren(
    clipboard.nodes.filter((n) => !isAllowed(n.type!)),
    clipboard.nodes,
    clipboard.edges
  );

  const sourceNodes = clipboard.nodes.filter((n) => !droppedIds.has(n.id));
  if (sourceNodes.length === 0) return [[], []];

  const minX = Math.min(...sourceNodes.map((n) => n.position.x));
  const minY = Math.min(...sourceNodes.map((n) => n.position.y));

  const target = position ?? { x: minX + 50, y: minY + 50 };

  const newIds = new Map(sourceNodes.map((n) => [n.id, getNodeId()]));

  const nodes = sourceNodes.map((n) => ({
    id: newIds.get(n.id)!,
    type: n.type,
    position: {
      x: n.position.x - minX + target.x,
      y: n.position.y - minY + target.y,
    },
    data: structuredClone(n.data),
    selected: true,
  }));

  const edges = clipboard.edges
    .filter((e) => newIds.has(e.source) && newIds.has(e.target))
    .map((e) => ({
      ...e,
      id: getEdgeId(),
      source: newIds.get(e.source)!,
      target: newIds.get(e.target)!,
    }));

  return [nodes, edges];
}

// Returns the IDs of the given nodes plus all children owned by them, e.g. the
// condition items of a condition.
function withOwnedChildren(
  roots: Node<NodeData>[],
  nodes: Node<NodeData>[],
  edges: Edge[]
): Set<string> {
  const nodesById = new Map(nodes.map((n) => [n.id, n]));
  const ids = new Set<string>();

  const add = (id: string) => {
    const node = nodesById.get(id);
    if (!node || ids.has(id)) return;
    ids.add(id);

    if (getNodeValues(node.type!).ownsChildren) {
      edges.filter((e) => e.source === id).forEach((e) => add(e.target));
    }
  };
  roots.forEach((n) => add(n.id));

  return ids;
}
