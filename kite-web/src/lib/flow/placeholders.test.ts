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

  it("lists all picked options of modal inputs in the modal's sub-flow", () => {
    const input = (custom_id: string, type: string, max_values?: number) => ({
      type: "label",
      label: custom_id,
      components: [
        {
          type,
          custom_id,
          max_values,
          options: [{ label: "A" }, { label: "B" }],
        },
      ],
    });
    const modalNodes = [
      node("entry", "entry_command"),
      node("modal", "suspend_response_modal", {
        modal_data: {
          title: "Form",
          components: [
            input("single", "string_select"),
            input("multi", "string_select", 2),
            input("roles", "role_select", 3),
            input("extras", "checkbox_group"),
            input("size", "radio_group"),
          ],
        },
      }),
      node("log", "action_log"),
      node("msg", "action_response_create", {
        message_data: { components: [{ type: 1, components: [{ id: 7 }] }] },
      }),
      node("later", "action_log"),
    ];
    const modalEdges = [
      edge("entry", "modal"),
      edge("modal", "log"),
      edge("log", "msg"),
      edge("msg", "later", "component_7"),
    ];
    const modalInputs = (nodeId: string) =>
      getAvailablePlaceholders(nodeId, modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.map((p) => p.value);

    expect(modalInputs("log")).toEqual([
      "input('single')",
      "input('multi')",
      "input('roles')",
      "input('extras')",
      "input('size')",
      "inputs('multi')",
      "inputs('roles')",
      "inputs('extras')",
    ]);
    expect(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.at(-1)?.label
    ).toBe("extras (Selected Values)");
    // Only listed until the next resume point.
    expect(modalInputs("later")).toEqual([
      "input('single')",
      "input('multi')",
      "input('roles')",
      "input('extras')",
      "input('size')",
    ]);
  });

  it("lists the reaction emoji for discord events", () => {
    const groups = getAvailablePlaceholders(undefined, [], [], "event_discord");
    expect(groups.find((g) => g.label === "Emoji")?.placeholders).toEqual([
      { label: "Emoji", value: "emoji" },
      { label: "Emoji ID", value: "emoji.id" },
      { label: "Emoji Name", value: "emoji.name" },
      { label: "Emoji Mention", value: "emoji.mention" },
    ]);
  });
});
