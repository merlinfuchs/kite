import { describe, expect, it } from "vitest";
import { flowTestEffectMessage, flowTestEffectTitle } from "./flowTest";

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
