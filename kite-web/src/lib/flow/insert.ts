import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";
import { nodeSize, rowGap } from "./layout";
import { getBlockDefinition } from "../blocks";
import { canConnect, createEdge, createNode, getNodeOutputs } from "./nodes";
import { walkDownstream } from "./placeholders";

// Blocks can be inserted into any connection except the fixed ones to the
// blocks a block owns and the ones from options into the entry.
export function canInsertIntoEdge(
  edge: Pick<Edge, "type">,
  sourceType: string | undefined
) {
  return (
    edge.type !== "fixed" && !!sourceType && !sourceType.startsWith("option_")
  );
}

// The block the flow continues from after an inserted block. Blocks without
// outputs continue from the owned block their definition names, like After
// loop, and End flow doesn't continue.
function getContinuation([block, ...owned]: Node<NodeData>[]) {
  if (getNodeOutputs(block).length > 0) return block;
  const type = getBlockDefinition(block.type)?.continues_from;
  return owned.find((n) => n.type === type);
}

// Blocks with named outputs continue from their default one, like the error
// handler, or else from the first, like yes of AI Check.
function getContinuationHandle(block: Node<NodeData>) {
  const outputs = getNodeOutputs(block);
  return outputs.includes("default") ? "default" : outputs[0];
}

// Inserts a new block into a connection, which then goes into the new block,
// and the new block continues to where it went. If there's no room between the
// two blocks, the blocks below are moved down to make some. Returns the new
// nodes and edges, with the inserted block selected, and its ID.
export function insertIntoEdge(
  type: string,
  edgeId: string,
  nodes: Node<NodeData>[],
  edges: Edge[]
): { nodes: Node<NodeData>[]; edges: Edge[]; id: string } | null {
  const edge = edges.find((e) => e.id === edgeId);
  const source = nodes.find((n) => n.id === edge?.source);
  const target = nodes.find((n) => n.id === edge?.target);
  if (
    !edge ||
    !source ||
    !target ||
    !canInsertIntoEdge(edge, source.type) ||
    !canConnect(source.type!, type) ||
    !canConnect(type, target.type!)
  ) {
    return null;
  }

  // Created at y 0, so the bottom of its lowest block is the height it needs.
  const [created, createdEdges] = createNode(
    type,
    { x: 0, y: 0 },
    { selected: true }
  );
  const height = Math.max(
    ...created.map((n) => n.position.y + nodeSize(n).height)
  );
  // Its width isn't known before it's rendered, so it's lined up with the
  // target's left edge.
  let x = target.position.x;

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
    x = (source.position.x + target.position.x) / 2;
    y = (source.position.y + target.position.y) / 2;
  }

  const moved = new Set(
    shift > 0 ? [target.id, ...walkDownstream([target.id], edges)] : []
  );

  const placed = created.map((n) => ({
    ...n,
    position: { x: n.position.x + x, y: n.position.y + y },
  }));
  const [inserted] = placed;
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
      createEdge(source, inserted, edge.sourceHandle),
      ...(continuation
        ? [
            createEdge(
              continuation,
              target,
              getContinuationHandle(continuation)
            ),
          ]
        : []),
    ],
    id: inserted.id,
  };
}
