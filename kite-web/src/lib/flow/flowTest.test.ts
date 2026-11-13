import { describe, expect, it } from "vitest";
import {
  canSetTestResult,
  flowTestEffectMessage,
  flowTestEffectTitle,
  flowTestNodeStatus,
  flowTestNoteText,
  parseTestValue,
  withFlowTestAction,
} from "./flowTest";

describe("flowTestEffectMessage", () => {
  it("shows a message in Discord's format", () => {
    const msg = flowTestEffectMessage({
      action: "create_message",
      data: {
        channel_id: "1",
        message: {
          content: "Pick one",
          embeds: [{ title: "Poll", description: null, color: 5814783 }],
          components: [
            {
              type: 1,
              components: [
                { type: 2, style: 1, label: "Yes", custom_id: "abc" },
                { type: 2, style: 5, label: "Docs", url: "https://kite.onl" },
              ],
            },
          ],
          flags: null,
          allowed_mentions: null,
        },
      },
    });

    expect(msg?.content).toBe("Pick one");
    expect(msg?.embeds[0].title).toBe("Poll");
    const row = msg?.components[0] as any;
    expect(row.components[0].flow_source_id).toBe("abc");
  });

  it("is null without a message", () => {
    expect(
      flowTestEffectMessage({ action: "add_reaction", data: { emoji: {} } })
    ).toBeNull();
  });
});

describe("flowTestEffectTitle", () => {
  it("names responses by their type", () => {
    expect(
      flowTestEffectTitle({ action: "respond", data: { type: "modal" } })
    ).toBe("Would show a modal");
    expect(flowTestEffectTitle({ action: "something_new" })).toBe(
      "Would something new"
    );
  });
});

describe("flowTestNodeStatus", () => {
  it("shows the worst of a block that ran more than once", () => {
    const status = flowTestNodeStatus({
      credits: 0,
      executions: [
        {
          trigger: "Command /loop",
          steps: [
            { node_id: "a", node_type: "action_message_create" },
            {
              node_id: "a",
              node_type: "action_message_create",
              effects: [{ action: "create_message", warnings: ["No access"] }],
            },
            { node_id: "b", node_type: "action_message_create", error: "x" },
            { node_id: "b", node_type: "action_message_create" },
            {
              node_id: "c",
              node_type: "action_http_request",
              notes: [
                "Not run, test runs don't send requests outside of Discord. Mock the block to set its result.",
              ],
            },
          ],
        },
      ],
    });
    expect(status).toEqual({ a: "warning", b: "error", c: "skipped" });
  });
});

describe("withFlowTestAction", () => {
  it("adds the action after the earlier ones", () => {
    const req = withFlowTestAction(
      { seed: 5, then: [{ click: "a" }] },
      { submit: { name: "x" } }
    );
    expect(req).toEqual({
      seed: 5,
      then: [{ click: "a" }, { submit: { name: "x" } }],
    });
  });
});

describe("flowTestNoteText", () => {
  it("rewrites the notes of the service for the editor", () => {
    const note =
      "Not run, test runs don't send requests outside of Discord. Mock the block to set its result.";
    expect(flowTestNoteText(note)).toContain("Set a test result");
    expect(canSetTestResult({ notes: [note] })).toBe(true);
    expect(flowTestNoteText("Waited 5m0s, which test runs skip")).toBe(
      "Waited 5m0s, which test runs skip"
    );
    expect(canSetTestResult({ notes: ["Waited 5m0s"] })).toBe(false);
  });
});

describe("parseTestValue", () => {
  it("keeps text that isn't JSON", () => {
    expect(parseTestValue("95")).toBe(95);
    expect(parseTestValue('{"xp": 3}')).toEqual({ xp: 3 });
    expect(parseTestValue("hello")).toBe("hello");
  });
});
