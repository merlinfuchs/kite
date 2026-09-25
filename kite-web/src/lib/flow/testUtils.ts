import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";

export function testNode(
  id: string,
  type: string,
  data: NodeData = {},
  position = { x: 0, y: 0 }
): Node<NodeData> {
  return { id, type, data, position };
}

export function testEdge(
  source: string,
  target: string,
  sourceHandle?: string
): Edge {
  return { id: `${source}-${target}`, source, target, sourceHandle };
}
