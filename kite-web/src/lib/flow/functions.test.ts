import { describe, expect, it } from "vitest";
import { FlowData } from "../types/flow.gen";
import {
  Command,
  EventListener,
  FlowFunction,
  Message,
} from "../types/wire.gen";
import {
  SharedFunction,
  describeFunctionUsages,
  getFunctionUsages,
  getMissingFunctions,
} from "./functions";
import { nodeCategories, orderSections } from "./categories";
import {
  collectReferencedResources,
  remapFlowReferences,
} from "./messageTemplates";

function flow(...nodes: { id: string; type: string; data?: object }[]) {
  return {
    nodes: nodes.map((n) => ({
      position: { x: 0, y: 0 },
      data: {},
      ...n,
    })),
    edges: [],
  } as unknown as FlowData;
}

const call = (id: string, functionId: string) => ({
  id,
  type: "action_function_call",
  data: { function_id: functionId },
});

const fnFlow = (...calls: string[]) =>
  flow(
    { id: "entry", type: "entry_function", data: { name: "Fn" } },
    ...calls.map((c, i) => call(`call${i}`, c))
  );

const shared = (id: string, ...calls: string[]): SharedFunction => ({
  id,
  flow_source: fnFlow(...calls),
});

describe("getMissingFunctions", () => {
  it("finds the bundled functions the app doesn't have, and the ones they call", () => {
    const command = flow(call("a", "check"), call("b", "existing"));
    const missing = getMissingFunctions(
      [command],
      [shared("check", "helper", "check"), shared("helper"), shared("unused")],
      new Set(["existing"])
    );
    expect(missing.map((f) => f.id)).toEqual(["check", "helper"]);
  });

  it("skips functions that weren't bundled", () => {
    expect(
      getMissingFunctions([flow(call("a", "gone"))], [], new Set())
    ).toEqual([]);
  });
});

describe("collectReferencedResources", () => {
  it("follows functions and message templates into each other", () => {
    const fns = [
      {
        id: "f1",
        flow_source: flow(
          { id: "entry", type: "entry_function" },
          {
            id: "msg",
            type: "action_message_create",
            data: { message_template_id: "m1" },
          }
        ),
      },
      { id: "f2", flow_source: fnFlow("f2") },
    ] as FlowFunction[];
    const messages = [
      {
        id: "m1",
        name: "Template",
        description: null,
        data: {},
        flow_sources: { button: flow(call("c", "f2")) },
      },
    ] as unknown as Message[];

    const res = collectReferencedResources(
      [flow(call("a", "f1"))],
      messages,
      fns
    );
    expect(res.functions.map((f) => f.id)).toEqual(["f1", "f2"]);
    expect(res.messages.map((m) => m.id)).toEqual(["m1"]);
  });

  it("leaves out excluded functions", () => {
    const fns = [{ id: "f1", flow_source: fnFlow("f1") }] as FlowFunction[];
    const res = collectReferencedResources([fnFlow("f1")], [], fns, [], ["f1"]);
    expect(res.functions).toEqual([]);
  });
});

describe("remapFlowReferences", () => {
  const command = flow(
    call("a", "imported"),
    call("b", "existing"),
    call("c", "gone")
  );

  it("points calls at imported copies and clears unknown functions", () => {
    const { flow: res, removed } = remapFlowReferences(command, {
      variableIds: new Set(),
      messageIds: new Set(),
      functionIds: new Set(["existing"]),
      functionIdMap: new Map([["imported", "copy"]]),
    });
    expect(res.nodes.map((n) => n.data.function_id)).toEqual([
      "copy",
      "existing",
      undefined,
    ]);
    expect(removed).toBe(1);
  });

  it("keeps function calls without a function mapping", () => {
    const { flow: res, removed } = remapFlowReferences(command, {
      variableIds: new Set(),
      messageIds: new Set(),
    });
    expect(res.nodes.map((n) => n.data.function_id)).toEqual([
      "imported",
      "existing",
      "gone",
    ]);
    expect(removed).toBe(0);
  });
});

describe("orderSections", () => {
  const titles = (
    category: "action" | "control_flow",
    context: "function" | "command"
  ) => orderSections(nodeCategories[category], context).map((s) => s.title);

  it("lists Call function under Actions right after Bot, in every flow", () => {
    for (const context of ["function", "command"] as const) {
      expect(titles("action", context).indexOf("Functions")).toBe(
        titles("action", context).indexOf("Bot") + 1
      );
    }
    const section = nodeCategories.action.find((s) => s.title === "Functions");
    expect(section?.nodeTypes).toEqual(["action_function_call"]);
  });

  it("lists Return first under Control Flow in functions", () => {
    expect(titles("control_flow", "function")[0]).toBe("Functions");
    expect(titles("control_flow", "command")[0]).not.toBe("Functions");
    const section = nodeCategories.control_flow.find(
      (s) => s.title === "Functions"
    );
    expect(section?.nodeTypes).toEqual(["control_function_return"]);
  });
});

describe("getFunctionUsages", () => {
  it("finds everything that calls the function, except itself", () => {
    const usages = getFunctionUsages("f1", {
      commands: [
        { id: "c1", name: "ban", flow_source: flow(call("a", "f1")) },
        { id: "c2", name: "ping", flow_source: flow(call("a", "f2")) },
      ] as Command[],
      eventListeners: [
        {
          id: "e1",
          description: "Welcome",
          type: "guild_member_add",
          flow_source: flow(call("a", "f1")),
        },
      ] as EventListener[],
      messages: [
        {
          id: "m1",
          name: "Panel",
          flow_sources: { button: flow(call("a", "f1")) },
        },
      ] as unknown as Message[],
      functions: [
        { id: "f1", name: "Self", flow_source: fnFlow("f1") },
        { id: "f3", name: "Other", flow_source: fnFlow("f1") },
      ] as FlowFunction[],
    });

    expect(usages.map((u) => u.label)).toEqual([
      "/ban",
      "event listener 'Welcome'",
      "message template 'Panel'",
      "function 'Other'",
    ]);
  });

  it("describes them in a sentence", () => {
    const usage = (label: string) => ({
      kind: "command" as const,
      id: label,
      label,
    });
    expect(describeFunctionUsages([usage("/a")])).toBe("/a");
    expect(describeFunctionUsages([usage("/a"), usage("/b")])).toBe(
      "/a and /b"
    );
    expect(describeFunctionUsages(["/a", "/b", "/c", "/d"].map(usage), 2)).toBe(
      "/a, /b and 2 more"
    );
  });
});
