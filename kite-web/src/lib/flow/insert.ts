import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";
import { nodeSize, rowGap } from "./layout";
import { createEdge, createNode, getNodeValues } from "./nodes";

// Blocks can be inserted into any connection except the fixed ones to the
// blocks a block owns and the ones from options into the entry.
export function canInsertIntoEdge(edge: Edge, nodes: Node<NodeData>[]) {
  const source = nodes.find((n) => n.id === edge.source);
  return (
    edge.type !== "fixed" && !!source && !source.type?.startsWith("option_")
  );
}

// The block the flow continues from after an inserted block. Blocks without
// outputs continue from what runs after a loop or from the branch of a
// condition, and End flow doesn't continue.
function getContinuation(created: Node<NodeData>[]) {
  const [block, ...owned] = created;
  if ((getNodeValues(block.type!).outputs ?? ["default"]).length > 0) {
    return block;
  }
  return owned.find(
    (n) =>
      n.type === "control_loop_end" ||
      (n.type!.startsWith("control_condition_item_") &&
        n.type !== "control_condition_item_else")
  );
}

// Inserts a new block into a connection, which then goes into the new block,
// and the new block continues to where it went. If there's no room between the
// two blocks, the blocks below are moved down to make some. Returns the new
// nodes and edges and the ID of the inserted block.
export function insertIntoEdge(
  type: string,
  edgeId: string,
  nodes: Node<NodeData>[],
  edges: Edge[]
): { nodes: Node<NodeData>[]; edges: Edge[]; id: string } | null {
  const edge = edges.find((e) => e.id === edgeId);
  if (!edge || !canInsertIntoEdge(edge, nodes)) return null;
  const source = nodes.find((n) => n.id === edge.source)!;
  const target = nodes.find((n) => n.id === edge.target);
  if (!target) return null;

  // Created at y 0, so the bottom of its lowest block is the height it needs.
  const [created, createdEdges] = createNode(type, { x: 0, y: 0 });
  const block = created[0];
  const height = Math.max(
    ...created.map((n) => n.position.y + nodeSize(n).height)
  );
  // Its width isn't known before it's rendered, so it's lined up with the
  // target's left edge.
  const x = target.position.x;

  const sourceBottom = source.position.y + nodeSize(source).height;
  const room = target.position.y - sourceBottom;
  let y: number;
  let shift = 0;
  if (room >= height + 2 * rowGap) {
    y = sourceBottom + (room - height) / 2;
  } else if (target.position.y > source.position.y) {
    y = sourceBottom + rowGap;
    shift = y + height + rowGap - target.position.y;
  } else {
    // The connection goes up or sideways, so nothing is below to move.
    y = (source.position.y + target.position.y) / 2;
  }

  const moved = new Set<string>();
  if (shift > 0) {
    const queue = [target.id];
    for (const id of queue) {
      if (moved.has(id)) continue;
      moved.add(id);
      for (const e of edges) {
        if (e.source === id) queue.push(e.target);
      }
    }
  }

  const placed = created.map((n) => ({
    ...n,
    position: { x: n.position.x + x, y: n.position.y + y },
  }));
  const continuation = getContinuation(placed);

  return {
    nodes: [
      ...nodes.map((n) =>
        // Blocks above the target are only reached through a loop back.
        moved.has(n.id) && n.position.y >= target.position.y
          ? { ...n, position: { ...n.position, y: n.position.y + shift } }
          : n
      ),
      ...placed,
    ],
    edges: [
      ...edges.filter((e) => e.id !== edgeId),
      ...createdEdges,
      createEdge(source, placed[0], edge.sourceHandle),
      ...(continuation ? [createEdge(continuation, target)] : []),
    ],
    id: block.id,
  };
}
