import { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { copyFlowNodes, pasteFlowNodes } from "./clipboard";
import { NodeData } from "./dataSchema";
import { applyFlowEdits } from "./edits";
import {
  applyGroupVisibility,
  fitGroupToMembers,
  getGroupData,
  getGroupMemberIds,
  getGroupProxyEdges,
  getHiddenNodeIds,
  groupHeaderHeight,
  groupNodeType,
  groupSelectedNodes,
  isGroupNode,
  setGroupCollapsed,
  storeGroupSizes,
  withGroupMembers,
  withGroupSizes,
} from "./groups";
import { getFlowLayout } from "./layout";
import { serializeFlow } from "./serialize";
import { validateFlow } from "./validate";

function block(id: string, x: number, y: number): Node<NodeData> {
  return {
    id,
    type: "action_log",
    position: { x, y },
    measured: { width: 200, height: 60 },
    data: { log_level: "info", log_message: "hi" },
  };
}

function note(id: string, x: number, y: number): Node<NodeData> {
  return {
    id,
    type: "note",
    position: { x, y },
    width: 100,
    height: 60,
    data: { note_text: "hi" },
  };
}

function edge(source: string, target: string): Edge {
  return { id: `${source}-${target}`, source, target };
}

// A group as it's saved: only its data and position.
const saved: Node<NodeData> = {
  id: "g",
  type: groupNodeType,
  position: { x: 0, y: 0 },
  data: {
    group_label: "Tickets",
    group_color: "red",
    group_width: 400,
    group_height: 400,
  },
};
const [group] = withGroupSizes([saved]);

describe("flow groups", () => {
  it("get their saved size and are drawn below the blocks", () => {
    expect(group).toMatchObject({ width: 400, height: 400 });
    expect(group.zIndex).toBeLessThan(0);
    expect(getGroupData(group)).toMatchObject({
      label: "Tickets",
      color: "#ef4444",
      collapsed: false,
    });
  });

  it("contain the blocks and notes whose center is inside them", () => {
    const nodes = [
      group,
      block("in", 50, 100),
      note("n", 300, 300),
      block("out", 500, 100),
    ];
    expect(getGroupMemberIds(group, nodes)).toEqual(["in", "n"]);
    expect(withGroupMembers(["g"], nodes).sort()).toEqual(["g", "in", "n"]);
  });

  it("hide their blocks when collapsed and show them when expanded", () => {
    const nodes = [group, block("in", 50, 100), block("out", 500, 100)];

    const collapsed = setGroupCollapsed(nodes, "g", true);
    expect(collapsed[0].height).toBe(groupHeaderHeight);
    expect(getHiddenNodeIds(collapsed)).toEqual(new Set(["in"]));
    expect(
      applyGroupVisibility(collapsed)
        .filter((n) => n.hidden)
        .map((n) => n.id)
    ).toEqual(["in"]);

    const expanded = setGroupCollapsed(collapsed, "g", false);
    expect(expanded[0].height).toBe(400);
    expect(getHiddenNodeIds(expanded).size).toBe(0);
  });

  it("don't hide blocks moved over them while collapsed", () => {
    const collapsed = setGroupCollapsed(
      [group, block("in", 50, 100)],
      "g",
      true
    );
    const nodes = [...collapsed, block("later", 50, 200)];
    expect(getHiddenNodeIds(nodes)).toEqual(new Set(["in"]));
  });

  it("deselect the blocks they hide", () => {
    const nodes = [group, { ...block("in", 50, 100), selected: true }];
    const collapsed = setGroupCollapsed(nodes, "g", true);
    expect(collapsed[1].selected).toBe(false);
  });

  it("are saved with their size and collapsed state, and loaded back", () => {
    const nodes = applyGroupVisibility(
      setGroupCollapsed(
        [
          { ...group, width: 420 },
          block("in", 50, 100),
          block("out", 500, 100),
        ],
        "g",
        true
      )
    );
    const stored = storeGroupSizes(nodes);

    expect(stored.every((n) => n.hidden === undefined)).toBe(true);
    expect(stored[0]).not.toHaveProperty("zIndex");
    expect(stored[0].data).toEqual({
      group_label: "Tickets",
      group_color: "red",
      group_collapsed: true,
      group_width: 420,
      group_height: 400,
    });

    // As the service returns it.
    const loaded = withGroupSizes(
      stored.map(({ id, type, position, data }) => ({
        id,
        type,
        position,
        data,
      }))
    );
    expect(loaded[0]).toMatchObject({ width: 420, height: groupHeaderHeight });
    expect(getHiddenNodeIds(loaded)).toEqual(new Set(["in"]));
  });

  it("draw the connections of hidden blocks to the group", () => {
    const nodes = setGroupCollapsed(
      [
        group,
        block("a", 50, 100),
        block("b", 50, 250),
        block("before", 500, 0),
        block("after", 500, 500),
      ],
      "g",
      true
    );
    const edges = [
      edge("before", "a"),
      edge("a", "b"),
      edge("b", "after"),
      { ...edge("b", "after"), id: "duplicate" },
    ];

    expect(getGroupProxyEdges(nodes, edges)).toEqual([
      {
        id: "before:->g:",
        source: "before",
        sourceHandle: null,
        target: "g",
        targetHandle: null,
      },
      {
        id: "g:->after:",
        source: "g",
        sourceHandle: null,
        target: "after",
        targetHandle: null,
      },
    ]);
  });

  it("can be made from the selected blocks", () => {
    const nodes = [
      { ...block("a", 0, 0), selected: true },
      { ...block("b", 300, 200), selected: true },
      block("c", 1000, 1000),
    ];
    const res = groupSelectedNodes(nodes)!;

    expect(isGroupNode(res.nodes[0])).toBe(true);
    expect(res.group.id).toMatch(/^group_/);
    expect(res.group.selected).toBe(true);
    expect(getGroupMemberIds(res.group, res.nodes)).toEqual(["a", "b"]);
    expect(res.nodes.filter((n) => n.selected)).toEqual([res.group]);
    expect(groupSelectedNodes([block("a", 0, 0)])).toBeNull();
  });

  it("fit around their blocks", () => {
    const nodes = [group, block("a", 100, 150)];
    const [fitted] = fitGroupToMembers(nodes, "g");
    expect(getGroupMemberIds(fitted, nodes)).toEqual(["a"]);
    expect(fitted.width).toBeLessThan(400);
  });

  it("are copied with their blocks and pasted expanded", () => {
    const nodes = setGroupCollapsed(
      [group, block("a", 50, 100), block("out", 500, 100)],
      "g",
      true
    );
    const clipboard = copyFlowNodes(nodes, [], ["g"])!;
    expect(clipboard.nodes.map((n) => n.id).sort()).toEqual(["a", "g"]);

    const [pasted] = pasteFlowNodes(clipboard, { x: 0, y: 0 });
    const loaded = withGroupSizes(pasted);
    const pastedGroup = loaded.find(isGroupNode)!;
    expect(getGroupData(pastedGroup).collapsed).toBe(false);
    expect(getGroupMemberIds(pastedGroup, loaded)).toHaveLength(1);
  });

  it("keep their blocks and notes in them when formatting", () => {
    const nodes = [
      group,
      block("entry", -600, -600),
      block("a", 50, 100),
      block("b", 50, 250),
      note("n", 250, 60),
      block("after", 900, 900),
    ];
    const edges = [edge("entry", "a"), edge("a", "b"), edge("b", "after")];

    const laidOut = getFlowLayout(nodes, edges);
    const g = laidOut.find((n) => n.id === "g")!;
    expect(getGroupMemberIds(g, laidOut).sort()).toEqual(["a", "b", "n"]);
    expect(getGroupMemberIds(g, laidOut)).not.toContain("after");
  });

  it("move hidden blocks along with the collapsed group when formatting", () => {
    const nodes = setGroupCollapsed(
      [group, block("entry", -600, -600), block("a", 50, 100)],
      "g",
      true
    );
    const laidOut = getFlowLayout(nodes, [edge("entry", "a")]);

    const before = nodes.find((n) => n.id === "g")!;
    const after = laidOut.find((n) => n.id === "g")!;
    const dx = after.position.x - before.position.x;
    const dy = after.position.y - before.position.y;
    expect(laidOut.find((n) => n.id === "a")!.position).toEqual({
      x: 50 + dx,
      y: 100 + dy,
    });
  });

  it("move notes the arranged blocks cover out of the way", () => {
    const nodes = [
      block("entry", 0, 0),
      block("a", 900, 900),
      note("covered", 0, 150),
      note("free", 2000, -500),
    ];
    const laidOut = getFlowLayout(nodes, [edge("entry", "a")]);
    const covered = laidOut.find((n) => n.id === "covered")!;
    const blocksRight = Math.max(
      ...laidOut
        .filter((n) => n.type === "action_log")
        .map((n) => n.position.x + 200)
    );
    expect(covered.position.x).toBeGreaterThan(blocksRight);
    expect(laidOut.find((n) => n.id === "free")!.position).toEqual({
      x: 2000,
      y: -500,
    });
  });

  it("aren't blocks for validation or the AI", () => {
    const entry: Node<NodeData> = {
      id: "entry",
      type: "entry_command",
      position: { x: 0, y: -200 },
      data: { name: "test", description: "Test" },
    };
    const nodes = [group, entry, block("a", 50, 100)];
    const edges = [edge("entry", "a")];

    expect(validateFlow(nodes, edges, "command")).toEqual(
      validateFlow([entry, block("a", 50, 100)], edges, "command")
    );

    const text = serializeFlow(nodes, edges, "command");
    expect(text).not.toContain("- g: flow_group");
    expect(text).toContain('- g "Tickets": a');

    const res = applyFlowEdits(
      { nodes, edges },
      [{ op: "connect", source: "a", target: "g" }],
      "command"
    );
    expect(res.edges).toEqual(edges);
  });
});
