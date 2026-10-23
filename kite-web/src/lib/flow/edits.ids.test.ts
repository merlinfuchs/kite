import { describe, expect, it } from "vitest";
import { applyFlowEdits } from "./edits";
import { testNode } from "./testUtils";

describe("applyFlowEdits", () => {
  it("gives added blocks IDs that aren't taken", () => {
    const entry = testNode("command1", "entry_command", {
      name: "test",
      description: "Test",
    });
    const log = testNode("log2", "action_log", {
      log_level: "info",
      log_message: "hi",
    });
    const res = applyFlowEdits(
      { nodes: [entry, log], edges: [] },
      [
        {
          op: "add_node",
          ref: "$log",
          type: "action_log",
          data: { log_level: "info", log_message: "hi" },
          after: "command1",
        },
      ],
      "command"
    );
    expect(res.refs.$log).toBe("log3");
    expect(res.nodes.map((n) => n.id)).toEqual(["command1", "log2", "log3"]);
  });
});
