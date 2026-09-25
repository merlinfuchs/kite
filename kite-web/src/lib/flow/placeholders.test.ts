import { describe, expect, it } from "vitest";
import { testEdge, testNode } from "./testUtils";
import { getAvailablePlaceholders } from "./placeholders";

const node = testNode;
const edge = testEdge;

const nodes = [
  node("entry", "entry_command"),
  node("arg", "option_command_argument", { name: "user" }),
  node("get", "action_user_get", { temporary_name: "target" }),
  node("msg", "action_response_create", {
    message_data: { components: [{ type: 1, components: [{ id: 7 }] }] },
  }),
  node("log", "action_log", { custom_label: "Logger" }),
];

const edges = [
  edge("arg", "entry"),
  edge("entry", "get"),
  edge("get", "msg"),
  edge("msg", "log", "component_7"),
];

const values = (nodeId?: string) =>
  getAvailablePlaceholders(nodeId, nodes, edges, "command").map((g) => [
    g.label,
    g.placeholders.map((p) => p.value),
  ]);

describe("getAvailablePlaceholders", () => {
  it("only lists flow wide placeholders without a block", () => {
    expect(values().map(([label]) => label)).toEqual([
      "Command",
      "User",
      "Server",
      "Channel",
      "App",
    ]);
    expect(values()[0]).toEqual(["Command", ["arg('user')"]]);
  });

  it("adds the results and variables of earlier blocks", () => {
    expect(values("msg").slice(5)).toEqual([
      ["Temporary Variables", ["var('target')"]],
      ["Node Results", ["result('get')"]],
    ]);
  });

  it("adds the original interaction after a resume point", () => {
    const labels = values("log").map(([label]) => label);
    expect(labels).toContain("Original User");
    expect(labels).not.toContain("Previous User");
    expect(values("log").at(-1)).toEqual([
      "Node Results",
      ["result('msg')", "result('get')"],
    ]);
  });
});
