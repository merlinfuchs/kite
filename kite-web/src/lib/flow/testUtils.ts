import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";

export function testNode(
  id: string,
  type: string,
  data: NodeData = {}
): Node<NodeData> {
  return { id, type, data, position: { x: 0, y: 0 } };
}

export function testEdge(
  source: string,
  target: string,
  sourceHandle?: string
): Edge {
  return { id: `${source}-${target}`, source, target, sourceHandle };
}
