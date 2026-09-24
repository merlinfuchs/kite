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
  const nodesById = new Map(nodes.map((n) => [n.id, n]));
  const copiedIds = new Set<string>();

  const add = (id: string) => {
    const node = nodesById.get(id);
    if (!node || copiedIds.has(id)) return;
    copiedIds.add(id);

    if (getNodeValues(node.type!).ownsChildren) {
      edges.filter((e) => e.source === id).forEach((e) => add(e.target));
    }
  };

  nodes
    .filter((n) => n.selected && !getNodeValues(n.type!).fixed)
    .forEach((n) => add(n.id));

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
// their top left corner is at the given position. Nodes rejected by
// isAllowed are dropped together with their edges and owned children.
export function pasteFlowNodes(
  clipboard: FlowClipboard,
  position: XYPosition,
  isAllowed: (type: string) => boolean = () => true
): [Node<NodeData>[], Edge[]] {
  const droppedIds = new Set<string>();
  const drop = (id: string) => {
    if (droppedIds.has(id)) return;
    droppedIds.add(id);

    const node = clipboard.nodes.find((n) => n.id === id);
    if (node && getNodeValues(node.type!).ownsChildren) {
      clipboard.edges
        .filter((e) => e.source === id)
        .forEach((e) => drop(e.target));
    }
  };
  clipboard.nodes.filter((n) => !isAllowed(n.type!)).forEach((n) => drop(n.id));

  const sourceNodes = clipboard.nodes.filter((n) => !droppedIds.has(n.id));
  if (sourceNodes.length === 0) return [[], []];

  const minX = Math.min(...sourceNodes.map((n) => n.position.x));
  const minY = Math.min(...sourceNodes.map((n) => n.position.y));

  const newIds = new Map(sourceNodes.map((n) => [n.id, getNodeId()]));

  const nodes = sourceNodes.map((n) => ({
    id: newIds.get(n.id)!,
    type: n.type,
    position: {
      x: n.position.x - minX + position.x,
      y: n.position.y - minY + position.y,
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
