import { afterEach, describe, expect, it, vi } from "vitest";
import { applyFlowEdits } from "./edits";
import { getOwnedChildTypes, getOwnerTypes } from "./nodes";
import { testNode } from "./testUtils";

describe("applyFlowEdits", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("gives added blocks IDs that aren't taken", () => {
    // Looking up owned block types creates blocks once per type, which
    // would use up the random values below.
    getOwnedChildTypes("action_log");
    getOwnerTypes("action_log");

    // The first ID is log_aaa, which is taken, then log_bbb.
    let calls = 0;
    vi.spyOn(Math, "random").mockImplementation(() =>
      calls++ < 3 ? 0 : 1 / 36
    );

    const entry = testNode("log_aaa", "entry_command", {
      name: "test",
      description: "Test",
    });
    const res = applyFlowEdits(
      { nodes: [entry], edges: [] },
      [
        {
          op: "add_node",
          ref: "$log",
          type: "action_log",
          data: { log_level: "info", log_message: "hi" },
          after: "log_aaa",
        },
      ],
      "command"
    );
    expect(res.refs.$log).toBe("log_bbb");
    expect(res.nodes.map((n) => n.id)).toEqual(["log_aaa", "log_bbb"]);
    expect(res.issues).toEqual([]);
  });
});
