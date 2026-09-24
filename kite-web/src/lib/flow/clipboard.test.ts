import { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import {
  copyFlowNodes,
  parseFlowClipboard,
  pasteFlowNodes,
} from "./clipboard";
import { NodeData } from "./dataSchema";

function node(
  id: string,
  type: string,
  selected = false,
  data: NodeData = {}
): Node<NodeData> {
  return { id, type, position: { x: 0, y: 0 }, data, selected };
}

function edge(source: string, target: string, type?: string): Edge {
  return { id: `${source}-${target}`, source, target, type };
}

describe("copyFlowNodes", () => {
  it("skips fixed nodes and returns null when nothing is copyable", () => {
    const nodes = [node("entry", "entry_command", true)];
    expect(copyFlowNodes(nodes, [])).toBeNull();
  });

  it("copies owned children and only edges between copied nodes", () => {
    const nodes = [
      node("entry", "entry_command", true),
      node("cond", "control_condition_compare", true),
      node("item", "control_condition_item_compare"),
      node("else", "control_condition_item_else"),
      node("after", "action_log"),
    ];
    const edges = [
      edge("entry", "cond"),
      edge("cond", "item", "fixed"),
      edge("cond", "else", "fixed"),
      edge("item", "after"),
    ];

    const clipboard = copyFlowNodes(nodes, edges)!;

    expect(clipboard.nodes.map((n) => n.id).sort()).toEqual([
      "cond",
      "else",
      "item",
    ]);
    expect(clipboard.edges.map((e) => e.id).sort()).toEqual([
      "cond-else",
      "cond-item",
    ]);
  });
});

describe("pasteFlowNodes", () => {
  it("assigns new ids, moves nodes and deep copies data", () => {
    const nodes = [
      { ...node("a", "action_log", true, { log_message: "hi" }) },
      { ...node("b", "action_log", true), position: { x: 100, y: 50 } },
    ];
    nodes[0].position = { x: 20, y: 10 };
    const clipboard = copyFlowNodes(nodes, [edge("a", "b")])!;

    const [newNodes, newEdges] = pasteFlowNodes(clipboard, { x: 500, y: 500 });

    expect(newNodes).toHaveLength(2);
    expect(newNodes.map((n) => n.id)).not.toContain("a");
    expect(newNodes.every((n) => n.selected)).toBe(true);
    expect(newNodes[0].position).toEqual({ x: 500, y: 500 });
    expect(newNodes[1].position).toEqual({ x: 580, y: 540 });
    expect(newNodes[0].data).not.toBe(clipboard.nodes[0].data);
    expect(newNodes[0].data).toEqual({ log_message: "hi" });

    expect(newEdges).toHaveLength(1);
    expect(newEdges[0].source).toBe(newNodes[0].id);
    expect(newEdges[0].target).toBe(newNodes[1].id);
  });

  it("drops disallowed nodes together with their owned children", () => {
    const nodes = [
      node("cond", "control_condition_compare", true),
      node("item", "control_condition_item_compare"),
      node("log", "action_log", true),
    ];
    const clipboard = copyFlowNodes(nodes, [edge("cond", "item", "fixed")])!;

    const [newNodes, newEdges] = pasteFlowNodes(
      clipboard,
      { x: 0, y: 0 },
      (type) => type !== "control_condition_compare"
    );

    expect(newNodes.map((n) => n.type)).toEqual(["action_log"]);
    expect(newEdges).toHaveLength(0);
  });
});

describe("parseFlowClipboard", () => {
  it("round trips and rejects other text", () => {
    const clipboard = copyFlowNodes([node("a", "action_log", true)], [])!;
    expect(parseFlowClipboard(JSON.stringify(clipboard))).toEqual(clipboard);
    expect(parseFlowClipboard("hello")).toBeNull();
    expect(parseFlowClipboard('{"nodes": []}')).toBeNull();
  });
});
