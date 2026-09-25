import { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { NodeData } from "./dataSchema";
import { applyFlowEdits, FlowEdit } from "./edits";
import { serializeFlow } from "./serialize";
import { testEdge as edge, testNode } from "./testUtils";

const node = testNode;

const entry = node("entry", "entry_command", {
  name: "test",
  description: "Test",
});

const log = (id: string, message = "hi", y = 0) =>
  node(
    id,
    "action_log",
    { log_level: "info", log_message: message },
    { x: 0, y }
  );

const logData = { log_level: "info", log_message: "hi" };

const addLog = (ref: string, after?: string, before?: string): FlowEdit => ({
  op: "add_node",
  ref,
  type: "action_log",
  data: logData,
  after,
  before,
});

function apply(nodes: Node<NodeData>[], edges: Edge[], edits: FlowEdit[]) {
  const res = applyFlowEdits({ nodes, edges }, edits, "command");
  const byId = (id: string) => res.nodes.find((n) => n.id === id);
  const connections = res.edges.map(
    (e) =>
      `${e.source}${e.sourceHandle ? `[${e.sourceHandle}]` : ""}->${e.target}`
  );
  return { ...res, byId, connections };
}

describe("applyFlowEdits", () => {
  it("adds blocks after others and below them", () => {
    const res = apply(
      [entry, log("a", "hi", 250)],
      [edge("entry", "a")],
      [addLog("$b", "a")]
    );

    const b = res.refs.$b;
    expect(res.issues).toEqual([]);
    expect(res.connections).toEqual(["entry->a", `a->${b}`]);
    expect(res.byId(b)?.position).toEqual({ x: 0, y: 500 });
    expect(res.edges[1].type).toBe("delete_button");
  });

  it("puts blocks into existing connections and moves the rest down", () => {
    const res = apply(
      [entry, log("a", "hi", 250), log("b", "hi", 500)],
      [edge("entry", "a"), edge("a", "b")],
      [addLog("$new", "entry", "a")]
    );

    const added = res.refs.$new;
    expect(res.issues).toEqual([]);
    expect(res.connections).toEqual(["a->b", `entry->${added}`, `${added}->a`]);
    expect(res.byId(added)?.position.y).toBe(250);
    expect(res.byId("a")?.position.y).toBe(500);
    expect(res.byId("b")?.position.y).toBe(750);
  });

  it("creates conditions with their branches", () => {
    const res = apply(
      [entry],
      [],
      [
        {
          op: "add_node",
          ref: "$check",
          type: "control_condition_compare",
          data: { condition_base_value: "{{user.id}}" },
          items: [
            { condition_item_mode: "equal", condition_item_value: "1" },
            { condition_item_mode: "equal", condition_item_value: "2" },
          ],
          after: "entry",
        },
        addLog("$one", "$check.item0"),
        addLog("$two", "$check.item1"),
        addLog("$other", "$check.else"),
      ]
    );

    expect(res.issues).toEqual([]);
    expect(res.byId(res.refs["$check.item1"])?.data).toEqual({
      condition_item_mode: "equal",
      condition_item_value: "2",
    });
    expect(
      res.edges.filter((e) => e.source === res.refs.$check).map((e) => e.type)
    ).toEqual(["fixed", "fixed", "fixed"]);
  });

  it("creates loops with their each and end blocks", () => {
    const res = apply(
      [entry],
      [],
      [
        {
          op: "add_node",
          ref: "$loop",
          type: "control_loop",
          data: { loop_count: "3" },
          after: "entry",
        },
        addLog("$each", "$loop.each"),
        {
          op: "add_node",
          ref: "$exit",
          type: "control_loop_exit",
          after: "$each",
        },
      ]
    );
    expect(res.issues).toEqual([]);
  });

  it("connects options into the entry and places them above it", () => {
    const res = apply(
      [node("entry", "entry_command", entry.data, { x: 0, y: 500 })],
      [],
      [
        {
          op: "add_node",
          ref: "$user",
          type: "option_command_argument",
          data: {
            name: "user",
            description: "User",
            command_argument_type: "user",
          },
        },
      ]
    );
    expect(res.issues).toEqual([]);
    expect(res.connections).toEqual([`${res.refs.$user}->entry`]);
    expect(res.byId(res.refs.$user)?.position).toEqual({ x: 0, y: 250 });
  });

  it("merges updated settings", () => {
    const message = node("msg", "action_response_create", {
      message_data: { content: "hi", allowed_mentions: { parse: ["users"] } },
      message_ephemeral: true,
    });
    const res = apply(
      [entry, message],
      [edge("entry", "msg")],
      [
        {
          op: "update_node",
          id: "msg",
          data: { message_data: { content: "bye" }, message_ephemeral: null },
        },
      ]
    );
    expect(res.byId("msg")?.data).toEqual({
      message_data: { content: "bye", allowed_mentions: { parse: ["users"] } },
    });
  });

  it("removes blocks and reconnects the ones after them", () => {
    const res = apply(
      [entry, log("a"), log("b")],
      [edge("entry", "a"), edge("a", "b")],
      [{ op: "remove_node", id: "a" }]
    );
    expect(res.connections).toEqual(["entry->b"]);
  });

  it("removes the blocks a condition owns with it", () => {
    const added = apply(
      [entry],
      [],
      [
        {
          op: "add_node",
          ref: "$check",
          type: "control_condition_compare",
          data: { condition_base_value: "a" },
          after: "entry",
        },
      ]
    );
    const res = applyFlowEdits(
      added,
      [{ op: "remove_node", id: added.refs.$check }],
      "command"
    );
    expect(res.nodes.map((n) => n.id)).toEqual(["entry"]);
    expect(res.edges).toEqual([]);
  });

  it("reports edits it can't apply and applies the rest", () => {
    const flow = [entry, log("a")];
    const res = apply(
      flow,
      [edge("entry", "a")],
      [
        { op: "remove_node", id: "entry" },
        { op: "update_node", id: "$missing", data: {} },
        addLog("$b", "a", "entry"),
        { op: "disconnect", source: "a", target: "entry" },
        addLog("$c", "a"),
      ]
    );

    expect(res.issues.map((i) => i.message)).toEqual([
      "Edit 1 (remove_node): 'entry' can't be removed on its own. Remove the block it belongs to instead.",
      "Edit 2 (update_node): There is no block '$missing'.",
      "Edit 3 (add_node): 'a' isn't connected to 'entry'.",
      "Edit 4 (disconnect): 'a' isn't connected to 'entry'.",
    ]);
    expect(res.nodes).toHaveLength(3);
    expect(res.connections).toEqual(["entry->a", `a->${res.refs.$c}`]);
  });

  it("reports what the validator finds", () => {
    const res = apply(
      [entry],
      [],
      [{ op: "add_node", ref: "$a", type: "action_log", after: "entry" }]
    );
    expect(res.issues.map((i) => i.message)).toEqual([
      "'Log Message' setting 'log_level': Required",
      "'Log Message' setting 'log_message': Required",
    ]);
  });
});

describe("applyFlowEdits edge cases", () => {
  it("names the default output of blocks with several outputs", () => {
    const handler = node("eh", "control_error_handler");
    const res = apply(
      [entry, handler],
      [edge("entry", "eh")],
      [addLog("$try", "eh")]
    );
    expect(res.edges.at(-1)?.sourceHandle).toBe("default");
    expect(res.issues).toEqual([]);
  });

  it("keeps the branches of a removed error handler and reconnects the default one", () => {
    const res = apply(
      [
        entry,
        node("eh", "control_error_handler"),
        log("t1"),
        log("t2"),
        log("e1"),
      ],
      [
        edge("entry", "eh"),
        edge("eh", "t1", "default"),
        edge("t1", "t2"),
        edge("eh", "e1", "error"),
      ],
      [{ op: "remove_node", id: "eh" }]
    );
    expect(res.nodes.map((n) => n.id)).toEqual(["entry", "t1", "t2", "e1"]);
    expect(res.connections).toEqual(["t1->t2", "entry->t1"]);
  });

  it("doesn't move button branches onto the direct path", () => {
    const message = node("msg", "action_response_create", {
      message_data: {
        components: [{ type: 1, components: [{ id: 7, type: 2, style: 1 }] }],
      },
    });
    const res = apply(
      [entry, message, log("next"), log("clicked")],
      [
        edge("entry", "msg"),
        edge("msg", "next"),
        edge("msg", "clicked", "component_7"),
      ],
      [{ op: "remove_node", id: "msg" }]
    );
    expect(res.connections).toEqual(["entry->next"]);
  });

  it("removes blocks with connections to missing blocks", () => {
    const res = apply(
      [entry, log("a")],
      [edge("entry", "a"), edge("ghost", "a")],
      [{ op: "remove_node", id: "a" }]
    );
    expect(res.issues.filter((i) => i.message.startsWith("Edit"))).toEqual([]);
    expect(res.nodes.map((n) => n.id)).toEqual(["entry"]);
  });

  it("rejects edits the editor doesn't allow", () => {
    const added = apply(
      [entry],
      [],
      [
        {
          op: "add_node",
          ref: "$check",
          type: "control_condition_compare",
          data: { condition_base_value: "a" },
          after: "entry",
        },
      ]
    );
    const res = applyFlowEdits(
      added,
      [
        {
          op: "disconnect",
          source: added.refs.$check,
          target: added.refs["$check.else"],
        },
        { op: "add_node", ref: "$else", type: "control_condition_item_else" },
        {
          op: "add_node",
          ref: "$loop",
          type: "control_loop",
          data: { loop_count: "1" },
          after: "entry",
          before: added.refs.$check,
        },
      ],
      "command"
    );
    expect(res.issues.map((i) => i.message).slice(0, 3)).toEqual([
      `Edit 1 (disconnect): '${added.refs["$check.else"]}' belongs to '${added.refs.$check}' and can't be disconnected. Remove it instead.`,
      "Edit 2 (add_node): 'control_condition_item_else' is created together with the block it belongs to.",
      `Edit 3 (add_node): 'control_loop' has no outputs, so nothing can come after it. Connect '${added.refs.$check}' to one of its branches instead.`,
    ]);
  });

  it("places blocks inserted before earlier added ones", () => {
    const res = apply(
      [entry, log("a", "hi", 250)],
      [edge("entry", "a")],
      [addLog("$b", "a"), addLog("$c", "a", "$b")]
    );
    expect(res.byId(res.refs.$c)?.position.y).toBe(500);
    expect(res.byId(res.refs.$b)?.position.y).toBe(750);
  });

  it("keeps appended blocks below the blocks moved for an insertion", () => {
    const res = apply(
      [entry, log("a", "hi", 250), log("b", "hi", 500)],
      [edge("entry", "a"), edge("a", "b")],
      [addLog("$n1", "entry", "a"), addLog("$n2", "b")]
    );
    expect(res.byId("b")?.position.y).toBe(750);
    expect(res.byId(res.refs.$n2)?.position.y).toBe(1000);
  });

  it("places blocks added before another one above it", () => {
    const res = apply(
      [entry, log("deep", "hi", 2000)],
      [],
      [
        {
          op: "add_node",
          ref: "$x",
          type: "action_log",
          data: logData,
          before: "deep",
        },
      ]
    );
    expect(res.byId(res.refs.$x)?.position.y).toBe(1750);
  });
});

describe("serializeFlow", () => {
  it("lists blocks in the order they run, without empty settings", () => {
    const arg = node("arg", "option_command_argument", {
      name: "user",
      description: "User",
      command_argument_type: "user",
      custom_label: "",
    });
    expect(
      serializeFlow(
        [log("b", "bye"), node("handler", "control_error_handler"), arg, entry],
        [
          edge("arg", "entry"),
          edge("entry", "handler"),
          edge("handler", "b", "error"),
        ],
        "command",
        ["b"]
      )
    ).toBe(
      [
        "Flow type: command",
        "",
        "Blocks:",
        '- entry entry_command {"name":"test","description":"Test"}',
        '- arg option_command_argument {"name":"user","description":"User","command_argument_type":"user"}',
        "- handler control_error_handler",
        '- b action_log (selected) {"log_level":"info","log_message":"bye"}',
        "",
        "Connections:",
        "- entry -> handler",
        "- arg -> entry",
        "- handler[error] -> b",
      ].join("\n")
    );
  });

  it("keeps arrays as they are", () => {
    const arg = node("arg", "option_command_argument", {
      name: "choice",
      command_argument_choices: [
        { name: "", value: "" },
        { name: "a", value: "a" },
      ],
    });
    expect(serializeFlow([arg], [], "command")).toContain(
      '"command_argument_choices":[{"name":"","value":""},{"name":"a","value":"a"}]'
    );
  });
});
