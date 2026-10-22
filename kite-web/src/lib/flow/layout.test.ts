import { describe, expect, it } from "vitest";
import { Edge, Node } from "@xyflow/react";
import { getLayoutedElements } from "./layout";

const node = (id: string, type: string, data: any = {}): Node => ({
  id,
  type,
  data,
  // Scrambled on purpose, the layout must not depend on it.
  position: { x: Math.random() * 1000, y: Math.random() * 1000 },
});

const edge = (source: string, target: string, sourceHandle?: string): Edge => ({
  id: `${source}-${target}`,
  source,
  target,
  sourceHandle,
});

function layout(nodes: Node[], edges: Edge[]) {
  const res = getLayoutedElements(nodes, edges);
  return new Map(res.nodes.map((n) => [n.id, n.position]));
}

describe("getLayoutedElements", () => {
  it("orders button and select menu flows like the message", () => {
    const components = [
      {
        type: 1,
        components: [
          { id: 1, type: 2, style: 1, label: "A" },
          { id: 2, type: 2, style: 1, label: "B" },
          { id: 3, type: 2, style: 1, label: "C" },
        ],
      },
      { type: 1, components: [{ id: 4, type: 3, options: [] }] },
    ];
    const nodes = [
      node("entry", "entry_command"),
      node("msg", "action_response_create", {
        message_data: { components },
      }),
      node("a", "action_log"),
      node("b", "action_log"),
      node("c", "action_log"),
      node("select", "action_log"),
      node("after", "action_log"),
    ];
    // Connected in a mixed up order.
    const edges = [
      edge("entry", "msg"),
      edge("msg", "after"),
      edge("msg", "c", "component_3"),
      edge("msg", "select", "component_4"),
      edge("msg", "a", "component_1"),
      edge("msg", "b", "component_2"),
    ];
    const p = layout(nodes, edges);
    const xs = ["a", "b", "c", "select", "after"].map((id) => p.get(id)!.x);
    expect(xs).toEqual([...xs].sort((x, y) => x - y));
    expect(p.get("a")!.y).toBe(p.get("after")!.y);
    expect(p.get("a")!.y).toBeGreaterThan(p.get("msg")!.y);
  });

  it("puts the error output left of the default output", () => {
    const nodes = [
      node("entry", "entry_command"),
      node("handler", "control_error_handler"),
      node("ok", "action_log"),
      node("failed", "action_log"),
    ];
    const edges = [
      edge("entry", "handler"),
      edge("handler", "ok", "default"),
      edge("handler", "failed", "error"),
    ];
    const p = layout(nodes, edges);
    expect(p.get("failed")!.x).toBeLessThan(p.get("ok")!.x);
  });

  it("puts the else branch last and the loop body first", () => {
    const nodes = [
      node("entry", "entry_command"),
      node("cond", "control_condition_compare"),
      node("else", "control_condition_item_else"),
      node("one", "control_condition_item_compare"),
      node("two", "control_condition_item_compare"),
      node("loop", "control_loop"),
      node("end", "control_loop_end"),
      node("each", "control_loop_each"),
    ];
    const edges = [
      edge("entry", "cond"),
      edge("cond", "else"),
      edge("cond", "one"),
      edge("cond", "two"),
      edge("one", "loop"),
      edge("loop", "end"),
      edge("loop", "each"),
    ];
    const p = layout(nodes, edges);
    expect(p.get("one")!.x).toBeLessThan(p.get("two")!.x);
    expect(p.get("two")!.x).toBeLessThan(p.get("else")!.x);
    expect(p.get("each")!.x).toBeLessThan(p.get("end")!.x);
  });

  it("places options above the entry and loose blocks beside the flow", () => {
    const nodes = [
      node("entry", "entry_command"),
      node("arg", "option_command_argument"),
      node("next", "action_log"),
      node("loose", "action_log"),
    ];
    const edges = [edge("arg", "entry"), edge("entry", "next")];
    const p = layout(nodes, edges);
    expect(p.get("arg")!.y).toBeLessThan(p.get("entry")!.y);
    expect(p.get("loose")!.x).toBeGreaterThan(p.get("next")!.x);
  });

  it("doesn't overlap blocks in the same row", () => {
    const nodes = [node("entry", "entry_command")];
    const edges: Edge[] = [];
    // A wide, uneven tree.
    for (let i = 0; i < 6; i++) {
      nodes.push(node(`l${i}`, "action_log"));
      edges.push(edge("entry", `l${i}`));
      for (let j = 0; j < i; j++) {
        nodes.push(node(`l${i}_${j}`, "action_log"));
        edges.push(edge(`l${i}`, `l${i}_${j}`));
      }
    }
    const p = layout(nodes, edges);
    const rows = new Map<number, number[]>();
    for (const pos of p.values()) {
      rows.set(pos.y, [...(rows.get(pos.y) ?? []), pos.x]);
    }
    for (const xs of rows.values()) {
      xs.sort((a, b) => a - b);
      for (let i = 1; i < xs.length; i++) {
        // Unmeasured blocks are 350 wide.
        expect(xs[i] - xs[i - 1]).toBeGreaterThanOrEqual(350);
      }
    }
  });

  it("handles cycles and connections to missing blocks", () => {
    const nodes = [node("a", "action_log"), node("b", "action_log")];
    const edges = [edge("a", "b"), edge("b", "a"), edge("a", "missing")];
    const p = layout(nodes, edges);
    expect(p.size).toBe(2);
    expect([...p.values()].every((v) => Number.isFinite(v.x))).toBe(true);
  });
});
