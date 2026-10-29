import { describe, expect, it } from "vitest";
import { getDeletedNodeIds, getNodeId, withOwnedNodes } from "./nodes";
import { testEdge as edge, testNode as node } from "./testUtils";

const nodes = [
  node("entry", "entry_command"),
  node("cond", "control_condition_compare"),
  node("item", "control_condition_item_compare"),
  node("else", "control_condition_item_else"),
  node("after", "action_log"),
  node("handler", "control_error_handler"),
  node("try", "action_log"),
];

const edges = [
  edge("entry", "cond"),
  edge("cond", "item"),
  edge("cond", "else"),
  edge("item", "after"),
  edge("entry", "handler"),
  edge("handler", "try", "default"),
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
