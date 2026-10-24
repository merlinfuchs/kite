import { describe, expect, it } from "vitest";
import {
  costsOnEveryFrequentEvent,
  formatFlowCredits,
  getFlowCredits,
  getTemplateCredits,
} from "./credits";
import { testEdge, testNode } from "./testUtils";
import { getTemplate } from "./templates";

describe("getFlowCredits", () => {
  it("adds up blocks that run one after another", () => {
    const nodes = [
      testNode("entry", "entry_command"),
      testNode("a", "action_response_create"),
      testNode("b", "action_member_ban"),
      testNode("c", "action_member_get"),
    ];
    const edges = [
      testEdge("entry", "a"),
      testEdge("a", "b"),
      testEdge("entry", "c"),
    ];
    expect(getFlowCredits(nodes, edges)).toEqual({
      min: 2,
      max: 2,
      variableLoops: false,
    });
  });

  it("takes one branch of a condition", () => {
    const nodes = [
      testNode("entry", "entry_command"),
      testNode("cond", "control_condition_compare"),
      testNode("yes", "control_condition_item_compare"),
      testNode("else", "control_condition_item_else"),
      testNode("a", "action_response_create"),
      testNode("b", "action_member_ban"),
      testNode("c", "action_response_create"),
    ];
    const edges = [
      testEdge("entry", "cond"),
      testEdge("cond", "yes"),
      testEdge("cond", "else"),
      testEdge("yes", "a"),
      testEdge("a", "b"),
      testEdge("else", "c"),
    ];
    expect(getFlowCredits(nodes, edges)).toMatchObject({ min: 1, max: 2 });

    // Without the else branch, nothing may run.
    expect(
      getFlowCredits(
        nodes.filter((n) => n.id !== "else"),
        edges.filter((e) => e.source !== "else" && e.target !== "else")
      )
    ).toMatchObject({ min: 0, max: 2 });
  });

  it("counts loops as often as they run", () => {
    const nodes = [
      testNode("entry", "entry_command"),
      testNode("loop", "control_loop", { loop_count: "3" }),
      testNode("each", "control_loop_each"),
      testNode("end", "control_loop_end"),
      testNode("a", "action_message_create"),
      testNode("b", "action_response_create"),
    ];
    const edges = [
      testEdge("entry", "loop"),
      testEdge("loop", "each"),
      testEdge("loop", "end"),
      testEdge("each", "a"),
      testEdge("end", "b"),
    ];
    expect(getFlowCredits(nodes, edges)).toEqual({
      min: 4,
      max: 4,
      variableLoops: false,
    });

    nodes[1] = testNode("loop", "control_loop", { loop_count: "{{x}}" });
    expect(getFlowCredits(nodes, edges)).toEqual({
      min: 1,
      max: 2,
      variableLoops: true,
    });
  });

  it("leaves out what buttons run", () => {
    const nodes = [
      testNode("entry", "entry_command"),
      testNode("msg", "action_response_create"),
      testNode("a", "action_member_role_add"),
    ];
    const edges = [
      testEdge("entry", "msg"),
      testEdge("msg", "a", "component_1"),
    ];
    expect(getFlowCredits(nodes, edges)).toMatchObject({ min: 1, max: 1 });
  });

  it("uses the cost of the AI model", () => {
    const nodes = [
      testNode("entry", "entry_command"),
      testNode("ai", "action_ai_chat_completion", {
        ai_chat_completion_data: { model: "medium" },
      }),
    ];
    expect(getFlowCredits(nodes, [testEdge("entry", "ai")]).max).toBe(20);
  });
});

describe("getTemplateCredits", () => {
  it("estimates every flow of a template", () => {
    const credits = getTemplateCredits(getTemplate("verification")!);
    expect(credits).toHaveLength(1);
    expect(credits[0]).toMatchObject({
      kind: "button",
      flowSourceId: "verify",
    });
    expect(formatFlowCredits(credits[0].credits)).toBe("up to 3 credits");
  });

  it("estimates the welcome message", () => {
    const credits = getTemplateCredits(getTemplate("welcomer")!);
    expect(credits[0]).toMatchObject({ kind: "event" });
    expect(formatFlowCredits(credits[0].credits)).toBe("1 credit");
  });
});

describe("costsOnEveryFrequentEvent", () => {
  const nodes = [
    testNode("entry", "entry_event", { event_type: "message_create" }),
    testNode("a", "action_message_reaction_create"),
  ];
  const edges = [testEdge("entry", "a")];

  it("flags listeners that pay for every message", () => {
    const credits = getFlowCredits(nodes, edges);
    expect(costsOnEveryFrequentEvent(nodes, edges, credits)).toBe(true);
  });

  it("trusts event filters", () => {
    const filtered = [...nodes, testNode("f", "option_event_filter")];
    const filteredEdges = [...edges, testEdge("f", "entry")];
    const credits = getFlowCredits(filtered, filteredEdges);
    expect(costsOnEveryFrequentEvent(filtered, filteredEdges, credits)).toBe(
      false
    );
  });

  it("ignores rare events", () => {
    const join = [
      testNode("entry", "entry_event", { event_type: "guild_member_add" }),
      nodes[1],
    ];
    const credits = getFlowCredits(join, edges);
    expect(costsOnEveryFrequentEvent(join, edges, credits)).toBe(false);
  });
});
