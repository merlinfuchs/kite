import { describe, expect, it } from "vitest";
import { createNode } from "./nodes";
import { getCachedMultiRunCounts, getMultiRunCounts } from "./multiRun";
import { testEdge, testNode } from "./testUtils";

const node = testNode;
const edge = testEdge;

const entry = node("entry", "entry_command", {
  name: "soulsplit",
  description: "Test",
});

const log = (id: string) =>
  node(id, "action_log", { log_level: "info", log_message: "hi" });

describe("getMultiRunCounts", () => {
  it("flags the user's diamond convergence without flagging downstream blocks", () => {
    const nodes = [entry, log("set"), log("get"), log("msg"), log("edit")];
    const edges = [
      edge("entry", "set"),
      edge("set", "get"),
      edge("get", "msg"),
      edge("entry", "msg"),
      edge("msg", "edit"),
    ];

    expect(getMultiRunCounts(nodes, edges)).toEqual(new Map([["msg", 2]]));
  });

  it("flags direct duplicate edges", () => {
    const nodes = [entry, log("msg")];
    const edges = [
      edge("entry", "msg"),
      { ...edge("entry", "msg"), id: "dup" },
    ];

    expect(getMultiRunCounts(nodes, edges)).toEqual(new Map([["msg", 2]]));
  });

  it("does not flag mutually exclusive branch merges", () => {
    const ifNode = node("cond", "control_if", { condition_base_value: "true" });
    expect(
      getMultiRunCounts(
        [entry, ifNode, log("a"), log("b"), log("after")],
        [
          edge("entry", "cond"),
          edge("cond", "a", "yes"),
          edge("cond", "b", "no"),
          edge("a", "after"),
          edge("b", "after"),
        ]
      )
    ).toEqual(new Map());

    const aiCheck = node("ai", "control_ai_check", {
      ai_decision_data: { question: "q", input: "i" },
    });
    expect(
      getMultiRunCounts(
        [entry, aiCheck, log("after")],
        [
          edge("entry", "ai"),
          edge("ai", "after", "yes"),
          edge("ai", "after", "no"),
          edge("ai", "after", "unsure"),
        ]
      )
    ).toEqual(new Map());

    const verify = node("verify", "control_webhook_verify");
    expect(
      getMultiRunCounts(
        [entry, verify, log("after")],
        [
          edge("entry", "verify"),
          edge("verify", "after", "valid"),
          edge("verify", "after", "invalid"),
        ]
      )
    ).toEqual(new Map());

    const errHandler = node("err", "control_error_handler");
    expect(
      getMultiRunCounts(
        [entry, errHandler, log("after")],
        [
          edge("entry", "err"),
          edge("err", "after", "default"),
          edge("err", "after", "error"),
        ]
      )
    ).toEqual(new Map());
  });

  it("respects condition_allow_multiple for match vs else branches", () => {
    const [cond, condEdges] = createNode("control_condition_compare", {
      x: 0,
      y: 0,
    });
    const condId = cond[0].id;
    const elseId = cond[1].id;
    const match1Id = cond[2].id;
    const match2 = node("match2", "control_condition_item_compare");

    const baseNodes = [entry, ...cond, match2, log("after")];
    const baseEdges = [
      edge("entry", condId),
      ...condEdges,
      edge(condId, "match2"),
    ];

    // Single-match condition: match1 + match2 + else converging is mutually exclusive.
    expect(
      getMultiRunCounts(baseNodes, [
        ...baseEdges,
        edge(match1Id, "after"),
        edge("match2", "after"),
        edge(elseId, "after"),
      ])
    ).toEqual(new Map());

    // Multi-match condition: match1 + else converging cannot both run.
    const multiCond = [
      entry,
      { ...cond[0], data: { condition_allow_multiple: true } },
      cond[1],
      cond[2],
      log("after"),
    ];
    expect(
      getMultiRunCounts(multiCond, [
        edge("entry", condId),
        ...condEdges,
        edge(match1Id, "after"),
        edge(elseId, "after"),
      ])
    ).toEqual(new Map());

    // Multi-match condition: match1 + match2 converging CAN both run.
    const multiCondBoth = [
      entry,
      { ...cond[0], data: { condition_allow_multiple: true } },
      cond[1],
      cond[2],
      match2,
      log("after"),
    ];
    expect(
      getMultiRunCounts(multiCondBoth, [
        ...baseEdges,
        edge(match1Id, "after"),
        edge("match2", "after"),
      ])
    ).toEqual(new Map([["after", 2]]));
  });

  it("handles loops properly", () => {
    const [loop, loopEdges] = createNode("control_loop", { x: 0, y: 0 });
    const [loopId, endId, eachId] = loop.map((n) => n.id);

    // Normal block inside loop each is NOT flagged as converging.
    expect(
      getMultiRunCounts(
        [entry, ...loop, log("body")],
        [edge("entry", loopId), ...loopEdges, edge(eachId, "body")]
      )
    ).toEqual(new Map());

    // Merging both each and end into the same block IS flagged.
    expect(
      getMultiRunCounts(
        [entry, ...loop, log("shared")],
        [
          edge("entry", loopId),
          ...loopEdges,
          edge(eachId, "shared"),
          edge(endId, "shared"),
        ]
      )
    ).toEqual(new Map([["shared", 2]]));
  });

  it("handles button resume branches independently", () => {
    const msg = node("panel", "action_response_create", {
      message_data: {
        content: "Pick",
        components: [
          {
            type: 1,
            components: [
              { id: 1, type: 2, style: 1, label: "One" },
              { id: 2, type: 2, style: 1, label: "Two" },
            ],
          },
        ],
      },
    });

    // Two separate buttons converging into the same block is NOT a multi-run.
    expect(
      getMultiRunCounts(
        [entry, msg, log("shared")],
        [
          edge("entry", "panel"),
          edge("panel", "shared", "component_1"),
          edge("panel", "shared", "component_2"),
        ]
      )
    ).toEqual(new Map());

    // Main flow default + button branch converging is NOT a multi-run.
    expect(
      getMultiRunCounts(
        [entry, msg, log("shared")],
        [
          edge("entry", "panel"),
          edge("panel", "shared"),
          edge("panel", "shared", "component_1"),
        ]
      )
    ).toEqual(new Map());

    // Diamond inside a single button branch IS flagged.
    expect(
      getMultiRunCounts(
        [entry, msg, log("prep"), log("after")],
        [
          edge("entry", "panel"),
          edge("panel", "prep", "component_1"),
          edge("prep", "after"),
          edge("panel", "after", "component_1"),
        ]
      )
    ).toEqual(new Map([["after", 2]]));
  });

  it("only flags the upstream convergence when a valid If merge sits below it", () => {
    const ifNode = node("cond", "control_if", { condition_base_value: "true" });
    const nodes = [
      entry,
      log("prep"),
      log("dup"),
      ifNode,
      log("yesBranch"),
      log("noBranch"),
      log("afterIf"),
    ];
    const edges = [
      edge("entry", "prep"),
      edge("prep", "dup"),
      edge("entry", "dup"),
      edge("dup", "cond"),
      edge("cond", "yesBranch", "yes"),
      edge("cond", "noBranch", "no"),
      edge("yesBranch", "afterIf"),
      edge("noBranch", "afterIf"),
    ];

    expect(getMultiRunCounts(nodes, edges)).toEqual(new Map([["dup", 2]]));
  });

  it("computes multiplicative counts for nested diamonds", () => {
    const nodes = [entry, log("a"), log("first"), log("b"), log("second")];
    const edges = [
      edge("entry", "a"),
      edge("a", "first"),
      edge("entry", "first"),
      edge("first", "b"),
      edge("b", "second"),
      edge("first", "second"),
    ];

    expect(getMultiRunCounts(nodes, edges)).toEqual(
      new Map([
        ["first", 2],
        ["second", 4],
      ])
    );
  });

  it("handles cycles without hanging or falsely flagging", () => {
    const nodes = [entry, log("a"), log("b")];
    const edges = [edge("entry", "a"), edge("a", "b"), edge("b", "a")];

    expect(getMultiRunCounts(nodes, edges)).toEqual(new Map());
  });

  it("ignores disconnected blocks not reachable from the entry", () => {
    const nodes = [entry, log("x"), log("y"), log("orphan")];
    const edges = [edge("x", "y"), edge("y", "orphan"), edge("x", "orphan")];

    expect(getMultiRunCounts(nodes, edges)).toEqual(new Map());
  });

  it("reuses cached results across node drags and invalidates when condition_allow_multiple changes", () => {
    const [cond, condEdges] = createNode("control_condition_compare", {
      x: 0,
      y: 0,
    });
    const match1Id = cond[2].id;
    const match2 = node("match2", "control_condition_item_compare");
    const edges = [
      edge("entry", cond[0].id),
      ...condEdges,
      edge(cond[0].id, "match2"),
      edge(match1Id, "after"),
      edge("match2", "after"),
    ];
    const nodes = [entry, ...cond, match2, log("after")];

    const first = getCachedMultiRunCounts(nodes, edges);
    expect(first).toEqual(new Map());

    const draggedNodes = nodes.map((n, i) =>
      i === nodes.length - 1 ? { ...n, position: { x: 100, y: 200 } } : n
    );
    const second = getCachedMultiRunCounts(draggedNodes, edges);
    expect(second).toBe(first);

    const toggledNodes = draggedNodes.map((n) =>
      n.id === cond[0].id
        ? { ...n, data: { ...n.data, condition_allow_multiple: true } }
        : n
    );
    const third = getCachedMultiRunCounts(toggledNodes, edges);
    expect(third).not.toBe(first);
    expect(third).toEqual(new Map([["after", 2]]));
  });
});
