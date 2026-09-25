import { describe, expect, it, vi } from "vitest";
import { applyFlowEdits } from "./edits";
import { getOwnedChildTypes } from "./nodes";
import { testNode } from "./testUtils";

let ids: string[] = [];
vi.mock("human-id", () => ({ humanId: () => ids.shift() ?? "later" }));

describe("applyFlowEdits", () => {
  it("gives added blocks IDs that aren't taken", () => {
    // Looking up owned block types creates blocks once per type, which
    // would use up the IDs below.
    getOwnedChildTypes("action_log");
    ids = ["entry", "fresh"];

    const entry = testNode("entry", "entry_command", {
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
          after: "entry",
        },
      ],
      "command"
    );
    expect(res.refs.$log).toBe("fresh");
    expect(res.nodes.map((n) => n.id)).toEqual(["entry", "fresh"]);
    expect(res.issues).toEqual([]);
  });
});
