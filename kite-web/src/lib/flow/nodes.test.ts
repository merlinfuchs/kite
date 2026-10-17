import { describe, expect, it } from "vitest";
import {
  canSpliceEdge,
  getActionableSelectedNodes,
  getDeletedNodeIds,
  getDisconnectedEdgeIds,
  getNodeId,
  getOwnerNodeId,
  getTopLevelSelectedNodes,
  withOwnedNodes,
} from "./nodes";
import { testEdge as edge, testNode as node } from "./testUtils";

const nodes = [
  node("entry", "entry_command"),
  node("cond", "control_condition_compare"),
  node("item", "control_condition_item_compare"),
  node("else", "control_condition_item_else"),
  node("after", "action_log"),
  node("handler", "control_error_handler"),
  node("try", "action_log"),
  node("opt", "option_command_argument"),
];

const edges = [
  edge("entry", "cond"),
  edge("cond", "item"),
  edge("cond", "else"),
  edge("item", "after"),
  edge("entry", "handler"),
  edge("handler", "try", "default"),
  edge("opt", "entry"),
];

describe("withOwnedNodes", () => {
  it("includes the blocks conditions and loops own, and nothing else", () => {
    expect([...withOwnedNodes(["cond", "handler"], nodes, edges)]).toEqual([
      "cond",
      "item",
      "else",
      "handler",
    ]);
  });
});

describe("getDeletedNodeIds", () => {
  it("deletes owned blocks that were selected together with their owner", () => {
    expect([
      ...getDeletedNodeIds(["cond", "item", "else"], nodes, edges),
    ]).toEqual(["cond", "item", "else"]);
  });

  it("keeps fixed blocks whose owner isn't deleted", () => {
    expect([
      ...getDeletedNodeIds(["entry", "else", "after"], nodes, edges),
    ]).toEqual(["after"]);
  });
});

describe("getNodeId", () => {
  it("names IDs after what the block is", () => {
    expect(getNodeId("action_message_create")).toMatch(/^message_[a-z0-9]{3}$/);
    expect(getNodeId("action_expression_evaluate")).toMatch(
      /^calc_[a-z0-9]{3}$/
    );
    expect(getNodeId("entry_command")).toMatch(/^command_[a-z0-9]{3}$/);
    expect(getNodeId("unknown")).toMatch(/^node_[a-z0-9]{3}$/);
    expect(getNodeId()).toMatch(/^node_[a-z0-9]{3}$/);
  });
});

describe("getTopLevelSelectedNodes & getActionableSelectedNodes", () => {
  it("only counts top-level blocks when a condition and its branches are selected", () => {
    const testNodes = [
      { ...nodes[0], selected: false },
      { ...nodes[1], selected: true },
      { ...nodes[2], selected: true },
      { ...nodes[3], selected: true },
      { ...nodes[4], selected: true },
    ];

    const topLevel = getTopLevelSelectedNodes(testNodes);
    expect(topLevel.map((n) => n.id)).toEqual(["cond", "after"]);

    const actionable = getActionableSelectedNodes(testNodes);
    expect(actionable.map((n) => n.id)).toEqual(["cond", "after"]);
  });

  it("excludes fixed entry blocks from actionable selection", () => {
    const testNodes = [
      { ...nodes[0], selected: true },
      { ...nodes[4], selected: true },
    ];

    const topLevel = getTopLevelSelectedNodes(testNodes);
    expect(topLevel.map((n) => n.id)).toEqual(["entry", "after"]);

    const actionable = getActionableSelectedNodes(testNodes);
    expect(actionable.map((n) => n.id)).toEqual(["after"]);
  });
});

describe("getOwnerNodeId", () => {
  it("finds the owner of condition items and else branches", () => {
    expect(getOwnerNodeId("item", nodes, edges)).toBe("cond");
    expect(getOwnerNodeId("else", nodes, edges)).toBe("cond");
    expect(getOwnerNodeId("cond", nodes, edges)).toBeNull();
    expect(getOwnerNodeId("after", nodes, edges)).toBeNull();
  });
});

describe("getDisconnectedEdgeIds", () => {
  it("collects non-fixed edges attached to the selected block and its owned children", () => {
    const disconnected = getDisconnectedEdgeIds(["cond"], nodes, edges);
    expect([...disconnected].sort()).toEqual(
      [edges[0].id, edges[3].id].sort()
    );
  });
});

describe("canSpliceEdge", () => {
  it("allows normal edges and blocks owned or option-to-entry edges", () => {
    expect(canSpliceEdge(edges[0], nodes)).toBe(true);
    expect(canSpliceEdge(edges[3], nodes)).toBe(true);
    expect(canSpliceEdge(edges[1], nodes)).toBe(false);
    expect(canSpliceEdge(edges[2], nodes)).toBe(false);
    expect(canSpliceEdge(edges[6], nodes)).toBe(false);
  });
});
