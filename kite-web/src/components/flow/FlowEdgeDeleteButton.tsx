import {
  BaseEdge,
  EdgeLabelRenderer,
  EdgeProps,
  getBezierPath,
  Position,
  useReactFlow,
  useStore,
} from "@xyflow/react";
import { XIcon } from "lucide-react";

const selfLoopOffset = 40;

export default function FlowEdgeDeleteButton({
  id,
  source,
  target,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  selected,
}: EdgeProps) {
  const { setEdges } = useReactFlow();
  // Only self-loops need the node bounds. Other edges select null, which never
  // changes, so they don't re-render when the source node does.
  const nodeRight = useStore((s) => {
    if (source !== target) return null;
    const node = s.nodeLookup.get(source);
    return node
      ? node.internals.positionAbsolute.x + (node.measured.width ?? 0)
      : null;
  });

  // A bezier from a node back to itself runs through the node, which hides the
  // delete button behind it. Route it around the right side instead.
  const pathArgs = {
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  };
  const [edgePath, labelX, labelY] =
    nodeRight !== null
      ? getSelfLoopPath({ ...pathArgs, nodeRight })
      : getBezierPath(pathArgs);

  const onEdgeClick = () => {
    if (selected) {
      setEdges((edges) => edges.filter((edge) => edge.id !== id));
    }
  };

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          stroke: selected ? "#6e6a95" : "#908dae",
          ...style,
        }}
      />
      <EdgeLabelRenderer>
        <div
          style={{
            position: "absolute",
            transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
            // everything inside EdgeLabelRenderer has no pointer events by default
            // if you have an interactive element, set pointer-events: all
            pointerEvents: "all",
          }}
          className="nodrag nopan cursor-pointer h-4 w-4 rounded-full flex items-center justify-center bg-muted"
          onClick={onEdgeClick}
        >
          <XIcon className="h-3 w-3 text-foreground" />
        </div>
      </EdgeLabelRenderer>
    </>
  );
}

function getSelfLoopPath({
  sourceX,
  sourceY,
  sourcePosition,
  targetX,
  targetY,
  nodeRight,
}: Pick<
  EdgeProps,
  "sourceX" | "sourceY" | "sourcePosition" | "targetX" | "targetY"
> & { nodeRight: number }): [string, number, number] {
  const sideX = Math.max(nodeRight, sourceX, targetX) + selfLoopOffset;
  const topY = targetY - selfLoopOffset / 2;
  const bottomY =
    sourcePosition === Position.Right ? sourceY : sourceY + selfLoopOffset / 2;

  const path = [
    `M ${sourceX},${sourceY}`,
    `L ${sourceX},${bottomY}`,
    `L ${sideX},${bottomY}`,
    `L ${sideX},${topY}`,
    `L ${targetX},${topY}`,
    `L ${targetX},${targetY}`,
  ].join(" ");

  return [path, sideX, (topY + bottomY) / 2];
}
