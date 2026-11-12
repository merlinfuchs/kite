import { describe, expect, it } from "vitest";
import { blockDefinitions } from "../blocks";
import { canInsertIntoEdge, insertIntoEdge } from "./insert";
import { testEdge as edge, testNode as node } from "./testUtils";

const at = (y: number) => ({ x: 0, y });

const nodes = [
  node("cooldown", "option_command_cooldown", {}, at(-200)),
  node("entry", "entry_command", {}, at(0)),
  node("handler", "control_error_handler", {}, at(210)),
  node("reply", "action_response_create", {}, at(420)),
  node("log", "action_log", {}, at(630)),
  node("cond", "control_condition_compare", {}, at(1500)),
  node("item", "control_condition_item_compare", {}, at(1700)),
];

const edges = [
  edge("cooldown", "entry"),
  edge("entry", "handler"),
  edge("handler", "reply", "error"),
  edge("reply", "log"),
  edge("log", "cond"),
  { ...edge("cond", "item"), type: "fixed" },
];

const byId = <T extends { id: string }>(list: T[]) =>
  Object.fromEntries(list.map((n) => [n.id, n]));

describe("canInsertIntoEdge", () => {
  it("allows connections other than fixed ones and options", () => {
    expect(canInsertIntoEdge(edges[1], "entry_command")).toBe(true);
    expect(canInsertIntoEdge(edges[0], "option_command_cooldown")).toBe(false);
    expect(canInsertIntoEdge(edges[5], "control_condition_compare")).toBe(
      false
    );
  });
});

describe("insertIntoEdge", () => {
  it("connects the new block in place of the connection", () => {
    const res = insertIntoEdge("action_log", "handler-reply", nodes, edges)!;

    expect(res.edges.find((e) => e.id === "handler-reply")).toBeUndefined();
    expect(res.edges).toContainEqual(
      expect.objectContaining({
        source: "handler",
        target: res.id,
        sourceHandle: "error",
      })
    );
    expect(res.edges).toContainEqual(
      expect.objectContaining({ source: res.id, target: "reply" })
    );
  });

  it("moves the blocks below down to make room", () => {
    const res = insertIntoEdge("action_log", "handler-reply", nodes, edges)!;
    const moved = byId(res.nodes);

    // 100 below the error handler, which is 140 high with its output labels,
    // and 100 above the reply.
    expect(moved[res.id].position.y).toBe(450);
    expect(moved.reply.position.y).toBe(660);
    expect(moved.log.position.y).toBe(870);
    expect(moved.cond.position.y).toBe(1740);
    expect(moved.item.position.y).toBe(1940);
    expect(moved.handler.position.y).toBe(210);
  });

  it("uses room that is already there", () => {
    const res = insertIntoEdge("action_log", "log-cond", nodes, edges)!;
    const moved = byId(res.nodes);

    expect(moved[res.id].position.y).toBe((740 + 1500 - 110) / 2);
    expect(moved.cond.position.y).toBe(1500);
  });

  it("continues after a loop and from the branch of a condition", () => {
    const loop = insertIntoEdge("control_loop", "reply-log", nodes, edges)!;
    const after = loop.nodes.find((n) => n.type === "control_loop_end")!;
    expect(loop.edges).toContainEqual(
      expect.objectContaining({ source: after.id, target: "log" })
    );

    const cond = insertIntoEdge(
      "control_condition_user",
      "reply-log",
      nodes,
      edges
    )!;
    const branch = cond.nodes.find(
      (n) => n.type === "control_condition_item_user"
    )!;
    expect(cond.edges).toContainEqual(
      expect.objectContaining({ source: branch.id, target: "log" })
    );
  });

  it("continues from the default or first named output", () => {
    const handler = insertIntoEdge(
      "control_error_handler",
      "reply-log",
      nodes,
      edges
    )!;
    expect(handler.edges).toContainEqual(
      expect.objectContaining({
        source: handler.id,
        target: "log",
        sourceHandle: "default",
      })
    );

    const check = insertIntoEdge(
      "control_ai_check",
      "reply-log",
      nodes,
      edges
    )!;
    expect(check.edges).toContainEqual(
      expect.objectContaining({
        source: check.id,
        target: "log",
        sourceHandle: "yes",
      })
    );
  });

  it("doesn't cover the target of a connection going sideways", () => {
    const side = [...nodes, node("side", "action_log", {}, { x: 500, y: 420 })];
    const res = insertIntoEdge("action_log", "reply-side", side, [
      ...edges,
      edge("reply", "side"),
    ])!;
    expect(byId(res.nodes)[res.id].position).toEqual({ x: 250, y: 420 });
  });

  it("doesn't insert options", () => {
    expect(
      insertIntoEdge("option_command_cooldown", "reply-log", nodes, edges)
    ).toBeNull();
  });

  it("knows where every block without outputs continues", () => {
    for (const block of blockDefinitions) {
      if (block.owns && block.outputs?.length === 0) {
        expect(block.owns, block.type).toContain(block.continues_from);
      }
    }
  });

  it("doesn't continue after End flow", () => {
    const res = insertIntoEdge("control_flow_end", "reply-log", nodes, edges)!;
    expect(res.edges.some((e) => e.target === "log")).toBe(false);
  });
});
