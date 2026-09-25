import { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { testEdge, testNode } from "./testUtils";
import { createNode } from "./nodes";
import { prepareTemplateFlow, getTemplates } from "./templates";
import { validateFlow } from "./validate";
import { ComponentData } from "../types/message.gen";

const node = testNode;
const edge = testEdge;

const entry = node("entry", "entry_command", {
  name: "test",
  description: "Test",
});

const arg = node("arg", "option_command_argument", {
  name: "user",
  description: "User",
  command_argument_type: "user",
});

const log = (id: string, message = "hi") =>
  node(id, "action_log", { log_level: "info", log_message: message });

function errors(
  nodes: Node<NodeData>[],
  edges: Edge[],
  context: FlowContextType = "command"
) {
  return validateFlow(nodes, edges, context)
    .filter((i) => i.severity === "error")
    .map((i) => i.message);
}

describe("validateFlow", () => {
  it("accepts the built-in templates", () => {
    for (const template of getTemplates()) {
      const inputs = Object.fromEntries(
        template.inputs.map((input) => [input.key, "123"])
      );
      const flows = [
        ...template.commands.map((c) => ({
          context: "command" as const,
          flow: c.flowSource(inputs),
        })),
        ...template.eventListeners.map((l) => ({
          context: "event_discord" as const,
          flow: l.flowSource(inputs),
        })),
      ];

      for (const { context, flow } of flows) {
        const { nodes, edges } = prepareTemplateFlow(flow);
        expect(validateFlow(nodes, edges, context)).toEqual([]);
      }
    }
  });

  it("needs exactly one entry of the flow's type", () => {
    expect(errors([log("a")], [])).toContain(
      "The flow needs exactly one entry block, but has 0."
    );
    expect(errors([entry], [], "event_discord")).toContain(
      "'Command' can't be used in this kind of flow."
    );
  });

  it("only connects options into the entry", () => {
    expect(
      errors([entry, arg, log("a")], [edge("arg", "a"), edge("a", "entry")])
    ).toEqual([
      "'Command Argument' can only be connected to the entry block.",
      "Only options can be connected into 'Command'.",
    ]);
  });

  it("checks outputs", () => {
    const handler = node("handler", "control_error_handler");
    expect(
      errors(
        [entry, handler, log("a"), log("b")],
        [
          edge("entry", "handler"),
          edge("handler", "a", "error"),
          edge("handler", "b", "default"),
        ]
      )
    ).toEqual([]);

    expect(
      errors(
        [entry, log("a"), log("b")],
        [edge("entry", "a"), edge("a", "b", "error")]
      )
    ).toEqual(["'Log Message' has no output 'error'."]);
  });

  it("accepts outputs of message components", () => {
    const message = node("msg", "action_response_create", {
      message_data: {
        components: [
          { type: 1, components: [{ id: 7, type: 2, style: 1, label: "a" }] },
        ],
      },
    });
    expect(
      errors(
        [entry, message, log("a")],
        [edge("entry", "msg"), edge("msg", "a", "component_7")]
      )
    ).toEqual([]);
    expect(
      errors(
        [entry, message, log("a")],
        [edge("entry", "msg"), edge("msg", "a", "component_8")]
      )
    ).toEqual(["'Create response message' has no output 'component_8'."]);
  });

  it("checks messages like the message editor", () => {
    const message = (components: ComponentData[]) =>
      node("msg", "action_response_create", {
        message_data: { components: [{ type: 1, components }] },
      });
    const button = (id?: number) => ({ id, type: 2, style: 1, label: "a" });
    const link = { type: 2, style: 5, label: "a", url: "https://example.com" };

    expect(errors([entry, message([button(1), button(2), link])], [])).toEqual(
      []
    );
    expect(errors([entry, message([button(1), button(1)])], [])).toEqual([
      "'Create response message' message components: two buttons or select menus have the id 1",
    ]);
    expect(
      errors([entry, message([{ id: 1, type: 2, style: 5, label: "a" }])], [])
    ).toEqual([
      "'Create response message' message components.0.components.0.url: Required",
    ]);
    expect(errors([entry, message([button()])], [])).toEqual([
      "'Create response message' message components: every button and select menu needs a number from 1 as id",
    ]);
  });

  it("doesn't check the message of blocks sending a template", () => {
    const message = node("msg", "action_response_create", {
      message_template_id: "t1",
      message_data: { components: [{ type: 1, components: [{ id: 1 }] }] },
    });
    expect(errors([entry, message], [])).toEqual([]);
  });

  it("checks the blocks owned by conditions and loops", () => {
    const [condition, conditionEdges] = createNode(
      "control_condition_compare",
      { x: 0, y: 0 }
    );
    condition[0].data = { condition_base_value: "a" };
    condition[2].data = { condition_item_mode: "equal" };
    const conditionId = condition[0].id;
    const elseId = condition[1].id;

    const valid = errors(
      [entry, ...condition],
      [edge("entry", conditionId), ...conditionEdges]
    );
    expect(valid).toEqual([]);

    expect(
      errors(
        [entry, ...condition.filter((n) => n.id !== elseId)],
        [
          edge("entry", conditionId),
          ...conditionEdges.filter((e) => e.target !== elseId),
        ]
      )
    ).toEqual([]);

    expect(
      errors(
        [entry, ...condition, log("a")],
        [edge("entry", conditionId), ...conditionEdges, edge(conditionId, "a")]
      )
    ).toEqual([
      "'Comparison Condition' has no outputs. Connect blocks to its branches instead.",
    ]);
  });

  it("only allows exiting a loop inside one", () => {
    const [loop, loopEdges] = createNode("control_loop", { x: 0, y: 0 });
    loop[0].data = { loop_count: "3" };
    const [loopId, endId, eachId] = loop.map((n) => n.id);
    const exit = node("exit", "control_loop_exit");
    expect(
      errors(
        [entry, ...loop, exit],
        [edge("entry", loopId), ...loopEdges, edge(eachId, "exit")]
      )
    ).toEqual([]);
    expect(
      errors(
        [entry, ...loop, exit],
        [edge("entry", loopId), ...loopEdges, edge(endId, "exit")]
      )
    ).toEqual(["'Exit loop' only works inside a loop."]);

    expect(
      errors(
        [entry, node("exit", "control_loop_exit")],
        [edge("entry", "exit")]
      )
    ).toEqual(["'Exit loop' only works inside a loop."]);
  });

  it("warns about blocks that never run", () => {
    expect(validateFlow([entry, arg, log("a")], [], "command")).toEqual([
      {
        severity: "warning",
        message:
          "'Command Argument' isn't connected to the entry block, so it has no effect.",
        nodeId: "arg",
      },
      {
        severity: "warning",
        message:
          "'Log Message' never runs, as it isn't connected to the entry block.",
        nodeId: "a",
      },
    ]);
  });

  it("checks settings", () => {
    expect(
      errors([entry, node("a", "action_log")], [edge("entry", "a")])
    ).toHaveLength(2);
  });

  it("checks that placeholders point at something that runs before", () => {
    const stored = node("get", "action_user_get", {
      user_target: "{{arg('user')}}",
      temporary_name: "target",
    });

    expect(
      errors(
        [
          entry,
          arg,
          stored,
          log("a", "{{var('target')}} {{result('get')}} {{nodes.get.result}}"),
        ],
        [edge("arg", "entry"), edge("entry", "get"), edge("get", "a")]
      )
    ).toEqual([]);

    expect(
      errors(
        [
          entry,
          stored,
          log("a", "{{var('target')}}"),
          node("calc", "action_expression_evaluate", {
            expression: "arg('missing') + 1",
          }),
        ],
        [edge("entry", "a"), edge("a", "get"), edge("get", "calc")]
      )
    ).toEqual([
      "'Get user' uses arg('user'), but the command has no argument named 'user'.",
      "'Calculate Value' uses arg('missing'), but the command has no argument named 'missing'.",
    ]);
  });

  it("warns about placeholders set by blocks that may run later", () => {
    const stored = node("get", "action_user_get", {
      user_target: "1",
      temporary_name: "target",
    });
    expect(
      validateFlow(
        [entry, stored, log("a", "{{var('target')}}")],
        [edge("entry", "get"), edge("entry", "a")],
        "command"
      )
    ).toEqual([
      {
        severity: "warning",
        message:
          "'Log Message' uses var('target'), which is set by a block that doesn't always run before it, so it may be empty.",
        nodeId: "a",
      },
    ]);
  });

  it("accepts any single placeholder in ID fields", () => {
    expect(
      errors(
        [
          entry,
          node("get", "action_user_get", {
            user_target: '{{ arg("user").id }}',
          }),
        ],
        [edge("entry", "get")]
      )
    ).toEqual([
      "'Get user' uses arg('user'), but the command has no argument named 'user'.",
    ]);
    expect(
      errors(
        [entry, node("get", "action_user_get", { user_target: "me" })],
        [edge("entry", "get")]
      )
    ).toEqual([
      "'Get user' setting 'user_target': Must be a number or ID, or a single {{ }} placeholder",
    ]);
  });

  it("doesn't allow connections into options", () => {
    expect(
      errors([entry, arg], [edge("arg", "entry"), edge("entry", "arg")])
    ).toEqual(["Nothing can be connected into 'Command Argument'."]);
  });

  it("handles blocks without data and duplicate connections", () => {
    expect(
      validateFlow(
        [entry, { id: "v", type: "action_voice_channel_leave" } as never],
        [edge("entry", "v"), { ...edge("entry", "v"), id: "dup" }],
        "command"
      )
    ).toEqual([
      {
        severity: "warning",
        message:
          "Two connections join the same blocks, so the second block runs twice.",
        edgeId: "dup",
      },
    ]);
  });

  it("finds modal inputs", () => {
    const modal = node("modal", "suspend_response_modal", {
      modal_data: {
        title: "Form",
        components: [
          { components: [{ custom_id: "name", label: "Name", style: 1 }] },
        ],
      },
    });
    expect(
      errors(
        [entry, modal, log("a", "{{input('name')}} {{input('age')}}")],
        [edge("entry", "modal"), edge("modal", "a")]
      )
    ).toEqual([
      "'Log Message' uses input('age'), but no modal before it has an input with the identifier 'age'.",
    ]);
  });

  it("counts earlier branches of loops and error handlers as running before", () => {
    const [loop, loopEdges] = createNode("control_loop", { x: 0, y: 0 });
    loop[0].data = { loop_count: "3" };
    const [loopId, endId, eachId] = loop.map((n) => n.id);
    const random = node("random", "action_random_generate", {
      random_min: "1",
      random_max: "10",
      temporary_name: "x",
    });
    expect(
      errors(
        [entry, ...loop, random, log("a", "{{var('x')}} {{result('random')}}")],
        [
          edge("entry", loopId),
          ...loopEdges,
          edge(eachId, "random"),
          edge(endId, "a"),
        ]
      )
    ).toEqual([]);

    const handler = node("handler", "control_error_handler");
    const get = node("get", "action_user_get", { user_target: "1" });
    expect(
      errors(
        [entry, handler, get, log("a", "{{result('get')}}")],
        [
          edge("entry", "handler"),
          edge("handler", "get"),
          edge("handler", "a", "error"),
        ]
      )
    ).toEqual([]);
  });

  it("only accepts arguments connected to the entry", () => {
    expect(
      errors([entry, arg, log("a", "{{arg('user')}}")], [edge("entry", "a")])
    ).toEqual([
      "'Log Message' uses arg('user'), but the command has no argument named 'user'.",
    ]);
  });

  it("checks the handles connections use", () => {
    const [loop, loopEdges] = createNode("control_loop", { x: 0, y: 0 });
    loop[0].data = { loop_count: "3" };
    expect(
      errors(
        [entry, ...loop],
        [
          { ...edge("entry", loop[0].id), targetHandle: "in" },
          ...loopEdges.map((e) => ({ ...e, sourceHandle: "each" })),
        ]
      )
    ).toEqual([
      "Connections can't target the input 'in'.",
      "'After loop' must be connected to the default output of 'Run a loop'.",
      "'Each loop iteration' must be connected to the default output of 'Run a loop'.",
    ]);
  });

  it("names every block an orphaned branch can belong to", () => {
    expect(
      errors(
        [entry, node("else", "control_condition_item_else")],
        [edge("entry", "else")]
      )
    ).toEqual([
      "'Else' can only be connected to the block it belongs to.",
      "'Else' must belong to exactly one of these blocks: 'Comparison Condition', 'User Condition', 'Channel Condition', 'Role Condition'.",
    ]);
  });

  it("reports block types it doesn't know", () => {
    expect(errors([entry, node("a", "toString")], [])).toEqual([
      "Unknown block type 'toString'.",
    ]);
  });
});
