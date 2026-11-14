import { describe, expect, it } from "vitest";
import {
  buildFlowTestRequest,
  flowTestEffectMessage,
  flowTestEffectTitle,
  initialFlowTestState,
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
                // Discord leaves out the label and animated.
                { type: 2, style: 2, emoji: { name: "👍" }, custom_id: "up" },
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
    expect(row.components[2].flow_source_id).toBe("up");
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

describe("buildFlowTestRequest", () => {
  const flow = { nodes: [], edges: [] };

  it("sends the arguments of a command by name", () => {
    const req = buildFlowTestRequest(
      { ...initialFlowTestState, options: { a1: " 5 ", a2: "", a3: "hi" } },
      {
        target: { type: "command", id: "c1" },
        flow,
        guildId: "g1",
        eventType: "",
        commandType: "chat_input",
        args: {
          a1: { name: "amount", type: "integer" },
          a2: { name: "reason", type: "string" },
          a3: { name: "text", type: "string" },
        },
      }
    );
    expect(req.options).toEqual({ amount: 5, text: "hi" });
    expect(req.flow_source).toBe(flow);
    expect(req.guild_id).toBe("g1");
  });

  it("rejects numbers Discord wouldn't send", () => {
    const build = (value: string) =>
      buildFlowTestRequest(
        { ...initialFlowTestState, options: { a1: value } },
        {
          target: { type: "command", id: "c1" },
          flow,
          guildId: null,
          eventType: "",
          commandType: "chat_input",
          args: { a1: { name: "amount", type: "integer" } },
        }
      );
    expect(() => build("5.5")).toThrow("amount must be a whole number.");
    expect(() => build("-")).toThrow();
    expect(build("-3").options).toEqual({ amount: -3 });
  });

  it("sends a webhook without a channel", () => {
    const req = buildFlowTestRequest(
      { ...initialFlowTestState, channelId: "ch1", webhookBody: "{}" },
      {
        target: { type: "event_listener", id: "e1" },
        flow,
        guildId: null,
        eventType: "webhook",
        commandType: "chat_input",
        args: {},
      }
    );
    expect(req.webhook).toEqual({ body: "{}" });
    expect(req.channel_id).toBeUndefined();
  });

  it("tests the flow of the dialog in the message", () => {
    const req = buildFlowTestRequest(
      { ...initialFlowTestState, valueIds: "1, 2" },
      {
        target: {
          type: "message",
          id: "m1",
          flowSourceId: "f1",
          getMessage: () => ({
            data: {
              components: [
                { type: 1, components: [{ type: 5, flow_source_id: "f1" }] },
              ],
            },
            flow_sources: { f1: { nodes: [], edges: [] }, f2: flow },
          }),
        },
        flow,
        guildId: null,
        eventType: "",
        commandType: "chat_input",
        args: {},
      }
    );
    expect(req.message?.flow_sources.f1).toBe(flow);
    expect(req.flow_source_id).toBe("f1");
    expect(req.values).toEqual(["1", "2"]);
  });
});
