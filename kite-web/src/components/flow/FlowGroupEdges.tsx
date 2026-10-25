import { NodeData } from "@/lib/flow/dataSchema";
import { getGroupProxyEdges, isGroupNode } from "@/lib/flow/groups";
import {
  getBezierPath,
  InternalNode,
  Node,
  Position,
  ReactFlowState,
  useStore,
  ViewportPortal,
} from "@xyflow/react";

interface Point {
  x: number;
  y: number;
  position: Position;
}

// Where an edge meets a node: the center of the handle's outer side, or of the
// bottom or top of a collapsed group.
function getEndpoint(
  node: InternalNode,
  type: "source" | "target",
  handleId: string | null
): Point | null {
  const { x, y } = node.internals.positionAbsolute;

  if (isGroupNode(node)) {
    const width = node.measured.width ?? node.width ?? 0;
    const height = node.measured.height ?? node.height ?? 0;
    return type === "source"
      ? { x: x + width / 2, y: y + height, position: Position.Bottom }
      : { x: x + width / 2, y, position: Position.Top };
  }

  const handles = node.internals.handleBounds?.[type] ?? [];
  const handle =
    (handleId !== null && handles.find((h) => h.id === handleId)) || handles[0];
  if (!handle) return null;

  const hx = x + handle.x;
  const hy = y + handle.y;
  switch (handle.position) {
    case Position.Top:
      return { x: hx + handle.width / 2, y: hy, position: handle.position };
    case Position.Bottom:
      return {
        x: hx + handle.width / 2,
        y: hy + handle.height,
        position: handle.position,
      };
    case Position.Left:
      return { x: hx, y: hy + handle.height / 2, position: handle.position };
    default:
      return {
        x: hx + handle.width,
        y: hy + handle.height / 2,
        position: handle.position,
      };
  }
}

// The paths are selected as one string, so the edges only re-render when one
// of them moves.
function selectPaths(s: ReactFlowState) {
  const nodes = s.nodes as Node<NodeData>[];
  if (!nodes.some((n) => n.hidden)) return "";

  const paths: string[] = [];
  for (const edge of getGroupProxyEdges(nodes, s.edges)) {
    const source = s.nodeLookup.get(edge.source);
    const target = s.nodeLookup.get(edge.target);
    if (!source || !target || source.hidden || target.hidden) continue;

    const from = getEndpoint(source, "source", edge.sourceHandle);
    const to = getEndpoint(target, "target", edge.targetHandle);
    if (!from || !to) continue;

    const [path] = getBezierPath({
      sourceX: from.x,
      sourceY: from.y,
      sourcePosition: from.position,
      targetX: to.x,
      targetY: to.y,
      targetPosition: to.position,
    });
    paths.push(path);
  }
  return paths.join("\n");
}

// Draws the connections of the blocks in collapsed groups to the groups, as
// the blocks themselves are hidden. They aren't edges of the flow, so they
// can't be selected or deleted.
export default function FlowGroupEdges() {
  const paths = useStore(selectPaths);
  if (!paths) return null;

  return (
    <ViewportPortal>
      <svg
        className="absolute left-0 top-0 overflow-visible pointer-events-none"
        style={{ zIndex: -1 }}
        width={1}
        height={1}
      >
        {paths.split("\n").map((path, i) => (
          <path
            key={i}
            d={path}
            fill="none"
            strokeWidth={1}
            strokeDasharray="5 4"
            style={{
              stroke: "var(--xy-edge-stroke, var(--xy-edge-stroke-default))",
            }}
          />
        ))}
      </svg>
    </ViewportPortal>
  );
}
