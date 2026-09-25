import { describe, expect, it } from "vitest";
import {
  emptyFlowHistory,
  FlowSnapshot,
  getFlowChangeKind,
  getFlowMergeKey,
  pushFlowHistory,
  redoFlowHistory,
  undoFlowHistory,
} from "./history";

function snapshot(id: string): FlowSnapshot {
  return {
    nodes: [{ id, type: "action_log", position: { x: 0, y: 0 }, data: {} }],
    edges: [],
  };
}

describe("flow history", () => {
  it("undoes and redoes in order", () => {
    const a = snapshot("a");
    const b = snapshot("b");
    const c = snapshot("c");

    let history = pushFlowHistory(emptyFlowHistory, a);
    history = pushFlowHistory(history, b);

    const [afterUndo, undone] = undoFlowHistory(history, c)!;
    expect(undone).toBe(b);
    expect(afterUndo).toEqual({ past: [a], future: [c] });

    const [afterRedo, redone] = redoFlowHistory(afterUndo, b)!;
    expect(redone).toBe(c);
    expect(afterRedo).toEqual({ past: [a, b], future: [] });
  });

  it("returns null when there is nothing to step to", () => {
    expect(undoFlowHistory(emptyFlowHistory, snapshot("a"))).toBeNull();
    expect(redoFlowHistory(emptyFlowHistory, snapshot("a"))).toBeNull();
  });

  it("drops the redo states on a new edit", () => {
    const history = { past: [], future: [snapshot("a")] };
    expect(pushFlowHistory(history, snapshot("b")).future).toEqual([]);
  });

  it("keeps at most 100 states", () => {
    let history = emptyFlowHistory;
    for (let i = 0; i < 150; i++) {
      history = pushFlowHistory(history, snapshot(`${i}`));
    }
    expect(history.past).toHaveLength(100);
    expect(history.past[0].nodes[0].id).toBe("50");
  });
});

describe("getFlowChangeKind", () => {
  it("ignores selection and measuring", () => {
    expect(getFlowChangeKind({ id: "a", type: "select", selected: true })).toBe(
      "ignore"
    );
    expect(getFlowChangeKind({ id: "a", type: "dimensions" })).toBe("ignore");
  });

  it("only counts positions once the drag has ended", () => {
    expect(
      getFlowChangeKind({ id: "a", type: "position", dragging: true })
    ).toBe("drag");
    expect(
      getFlowChangeKind({ id: "a", type: "position", dragging: false })
    ).toBe("edit");
  });

  it("counts adding, removing and replacing", () => {
    expect(getFlowChangeKind({ id: "a", type: "remove" })).toBe("edit");
    expect(
      getFlowChangeKind({ type: "add", item: snapshot("a").nodes[0] })
    ).toBe("edit");
  });
});

describe("getFlowMergeKey", () => {
  const item = snapshot("a").nodes[0];

  it("merges edits to the data of one node", () => {
    expect(getFlowMergeKey([{ id: "a", type: "replace", item }])).toBe(
      "data:a"
    );
  });

  it("merges keyboard moves", () => {
    expect(getFlowMergeKey([{ id: "a", type: "position" }])).toBe("move");
  });

  it("doesn't merge anything else", () => {
    expect(getFlowMergeKey([{ type: "add", item }])).toBeUndefined();
    expect(
      getFlowMergeKey([
        { id: "a", type: "replace", item },
        { id: "b", type: "replace", item },
      ])
    ).toBeUndefined();
  });
});
