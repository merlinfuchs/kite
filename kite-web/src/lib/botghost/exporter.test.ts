import { describe, expect, it } from "vitest";
import { helperCollectSource } from "./exporter";
import { helperBookmarklet } from "../importHelper";
import { parseBotGhostExport } from "./parse";

const collect = new Function(`${helperCollectSource}; return collect;`)() as (
  state: unknown
) => unknown;

describe("BotGhost helper", () => {
  it("copies only commands and custom events out of the dashboard state", () => {
    const state: Record<string, unknown> = {
      auth: { token: "secret", user: { name: "someone" } },
      data: {
        bot: {
          token: "secret",
          commands: {
            slash: {
              commands: [
                {
                  name: "ping",
                  description: "Pong",
                  options: [],
                  actions: [{ type: "plain_text", response: "Pong" }],
                },
              ],
            },
            custom_events: {
              events: [
                { name: "Welcome", type: "guildMemberAdd", actions: [] },
              ],
            },
          },
        },
      },
    };
    // The state of a dashboard can refer to itself.
    state.self = state;

    const text = JSON.stringify(collect(state));
    expect(text).not.toContain("secret");

    const parsed = parseBotGhostExport(text);
    expect(parsed.commands.map((c) => c.name)).toEqual(["ping"]);
    expect(parsed.events.map((e) => e.eventType)).toEqual(["guildMemberAdd"]);
  });

  it("is a bookmark that runs as it is", () => {
    expect(helperBookmarklet.startsWith("javascript:")).toBe(true);
    const source = decodeURIComponent(
      helperBookmarklet.slice("javascript:".length)
    );
    expect(() => new Function(source)).not.toThrow();
  });
});
