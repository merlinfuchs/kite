import { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { applyFlowCode, parseFlowCode, printFlowCode } from "./flowCode";
import { normalizeHandle } from "./nodes";
import { isNoteNode } from "./notes";
import { compact } from "./serialize";
import { getTemplates } from "./templateFiles";
import { testEdge as edge, testNode as node } from "./testUtils";

const entry = node("entry", "entry_command", {
  name: "test",
  description: "Test",
});

function connections(edges: Edge[]) {
  return edges
    .map(
      (e) =>
        `${e.source}${
          normalizeHandle(e.sourceHandle)
            ? `[${normalizeHandle(e.sourceHandle)}]`
            : ""
        }->${e.target}`
    )
    .sort();
}

function apply(nodes: Node<NodeData>[], edges: Edge[], code: string) {
  const res = applyFlowCode({ nodes, edges }, code, "command");
  return {
    ...res,
    errors: res.issues
      .filter((i) => i.severity === "error")
      .map((i) => i.message),
    connections: connections(res.edges),
    byId: (id: string) => res.nodes.find((n) => n.id === id),
  };
}

describe("printFlowCode and applyFlowCode", () => {
  const flows = getTemplates().flatMap((t) => [
    ...t.commands.map((c) => ({
      name: `${t.id}/${c.key}`,
      flow: c.flowSource,
      context: "command" as FlowContextType,
    })),
    ...t.eventListeners.map((l) => ({
      name: `${t.id}/${l.key}`,
      flow: l.flowSource,
      context: "event_discord" as FlowContextType,
    })),
  ]);

  it.each(flows)("keep $name as it is", ({ flow, context }) => {
    const nodes = flow.nodes as Node<NodeData>[];
    const edges = flow.edges as Edge[];
    const code = printFlowCode(nodes, edges, context);
    const res = applyFlowCode({ nodes, edges }, code, context);

    const parseErrors = res.issues.filter((i) => i.message.startsWith("Line"));
    expect(parseErrors).toEqual([]);
    const blocks = (ns: Node<NodeData>[]) =>
      ns
        .filter((n) => !isNoteNode(n))
        .map((n) => `${n.id} ${n.type} ${JSON.stringify(compact(n.data))}`)
        .sort();
    expect(blocks(res.nodes)).toEqual(blocks(nodes));
    expect(connections(res.edges)).toEqual(connections(edges));
    // Unchanged blocks keep their settings object, so they don't count as
    // changed by the AI.
    for (const n of res.nodes) {
      const before = nodes.find((b) => b.id === n.id);
      expect(n.data === before?.data).toBe(true);
    }
  });
});

describe("applyFlowCode", () => {
  it("adds blocks in order", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_log({ $id: "a", log_level: "info", log_message: "1" }),
        action_log({ $id: "b", log_level: "info", log_message: "2" }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual(["a->b", "entry->a"]);
  });

  it("connects options to the entry", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        option_command_argument({ $id: "arg", name: "who", description: "Who", command_argument_type: "user" }),
        action_log({ $id: "a", log_level: "info", log_message: "hi" }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual(["arg->entry", "entry->a"]);
  });

  it("uses the outputs of an If", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_if({ $id: "if", condition_base_value: "{{true}}" }, {
          yes: [action_log({ $id: "y", log_level: "info", log_message: "y" })],
          no: [action_log({ $id: "n", log_level: "info", log_message: "n" })],
        }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual(["entry->if", "if[no]->n", "if[yes]->y"]);
  });

  it("runs a block after an If after both of its outputs", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_if({ $id: "if", condition_base_value: "{{true}}" }, {
          yes: [action_log({ $id: "y", log_level: "info", log_message: "y" })],
        }),
        action_log({ $id: "after", log_level: "info", log_message: "after" }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual([
      "entry->if",
      "if[no]->after",
      "if[yes]->after",
      "if[yes]->y",
    ]);
    // It runs after the blocks of the output.
    const yes = res.edges.filter((e) => e.sourceHandle === "yes");
    expect(yes.map((e) => e.target)).toEqual(["y", "after"]);
    expect(
      printFlowCode(res.nodes, res.edges, "command").replace(/\s+/g, " ")
    ).toContain(
      'yes: [ action_log({ $id: "y", log_level: "info", log_message: "y" }), ], no: [], }), action_log({ $id: "after"'
    );
  });

  it("reads a list after an If as its yes output", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_if({ $id: "if", condition_base_value: "{{true}}" }, [
          action_log({ $id: "y", log_level: "info", log_message: "y" }),
        ]),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual(["entry->if", "if[yes]->y"]);
  });

  it("explains a block after a block without outputs", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_flow_end({}),
        action_log({ log_level: "info", log_message: "y" }),
      ])`
    );
    expect(res.errors[0]).toMatch(
      /^Line 3: control_flow_end\(\) has no outputs/
    );
    expect(res.nodes).toEqual([entry]);
  });

  it("creates conditions with their branches", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_condition_compare({ $id: "c", condition_base_value: "{{1}}" }, {
          branches: [
            control_condition_item_compare({ $id: "one", condition_item_mode: "equal", condition_item_value: "1" }, [
              action_log({ $id: "a", log_level: "info", log_message: "one" }),
            ]),
          ],
          else: [action_log({ $id: "b", log_level: "info", log_message: "else" })],
        }),
      ])`
    );
    expect(res.errors).toEqual([]);
    const elseId = res.nodes.find(
      (n) => n.type === "control_condition_item_else"
    )!.id;
    expect(res.connections).toEqual(
      ["c->one", `c->${elseId}`, "entry->c", "one->a", `${elseId}->b`].sort()
    );
  });

  it("reads conditions with their branches as a list", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_condition_compare({ $id: "c", condition_base_value: "{{1}}" }, [
          control_condition_item_compare({ $id: "one", condition_item_mode: "equal", condition_item_value: "1" }),
        ], [action_log({ $id: "b", log_level: "info", log_message: "else", custom_label: "" })]),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toContain("c->one");
    expect(res.byId("b")!.data).toEqual({
      log_level: "info",
      log_message: "else",
    });
  });

  it("keeps existing blocks and removes left out ones", () => {
    const a = node("a", "action_log", { log_level: "info", log_message: "a" });
    const b = node("b", "action_log", { log_level: "info", log_message: "b" });
    const res = apply(
      [entry, a, b],
      [edge("entry", "a"), edge("a", "b")],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_log({ $id: "a", log_level: "info", log_message: "a" }),
        action_log({ log_level: "info", log_message: "new" }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.byId("a")!.data).toBe(a.data);
    expect(res.byId("b")).toBeUndefined();
    expect(res.nodes).toHaveLength(3);
  });

  it("closes brackets the code left open", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_log({ $id: "a", log_level: "info", log_message: "a" ),
        action_log({ $id: "b", log_level: "info", log_message: "b" })`
    );
    expect(res.errors).toEqual([]);
    expect(res.connections).toEqual(["a->b", "entry->a"]);
  });

  it("copies a block written twice", () => {
    const a = node("a", "action_log", { log_level: "info", log_message: "a" });
    const res = apply(
      [entry, a],
      [edge("entry", "a")],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        control_if({ $id: "if", condition_base_value: "{{true}}" }, {
          yes: [action_log({ $id: "a", log_level: "info", log_message: "a" })],
          no: [action_log({ $id: "a", log_level: "info", log_message: "a" })],
        }),
      ])`
    );
    expect(res.errors).toEqual([]);
    expect(res.nodes.filter((n) => n.type === "action_log")).toHaveLength(2);
    expect(res.connections).toContain("if[yes]->a");
  });

  it("reads result() of a new block by its temporary_name", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_random_generate({ random_mode: "number", random_min: "1", random_max: "6", temporary_name: "roll" }),
        action_response_create({ $id: "r", message_data: { content: "{{result('roll')}}" } }),
      ])`
    );
    const id = res.nodes.find((n) => n.type === "action_random_generate")!.id;
    expect(res.byId("r")!.data.message_data?.content).toBe(
      `{{result('${id}')}}`
    );
  });

  it("reports the line of a syntax error", () => {
    const res = apply(
      [entry],
      [],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_log({ log_level: "info" log_message: "a" }),
      ])`
    );
    expect(res.errors[0]).toMatch(/^Line 2: Expected , or \}/);
  });

  it("refuses to change the type of a block", () => {
    const a = node("a", "action_log", { log_level: "info", log_message: "a" });
    const res = apply(
      [entry, a],
      [edge("entry", "a")],
      `entry_command({ $id: "entry", name: "test", description: "Test" }, [
        action_response_create({ $id: "a", message_data: { content: "hi" } }),
      ])`
    );
    expect(res.errors[0]).toMatch(/is a action_log, not/);
  });
});

describe("parseFlowCode", () => {
  it("reads strings with any quotes and escapes", () => {
    const [v] = parseFlowCode("['a\\'b', \"c\\nd\", `e\nf`, 'g\\u00e9']");
    expect(v).toMatchObject({
      kind: "array",
      items: [
        { value: "a'b" },
        { value: "c\nd" },
        { value: "e\nf" },
        { value: "gé" },
      ],
    });
  });

  it("strips a Markdown code block", () => {
    expect(parseFlowCode("```js\n[1]\n```")).toHaveLength(1);
  });
});
