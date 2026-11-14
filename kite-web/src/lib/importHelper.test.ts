import { describe, expect, it } from "vitest";
import {
  helperBookmarklet,
  helperConsoleSource,
  helperReadSource,
} from "./importHelper";
import { parseBotGhostExport } from "./botghost/parse";
import { parseInventorExport } from "./inventor/parse";

type Result = { data?: unknown; message: string };

// Runs the part of the helper that reads a dashboard, on a made-up page.
function run(page: {
  host: string;
  path?: string;
  fetch?: (url: string) => unknown;
  state?: unknown;
}): Promise<Result> {
  const fiber = {
    memoizedProps: { store: { getState: () => page.state } },
  };
  const body = { __reactContainer$x: fiber };
  const read = new Function(
    "location",
    "document",
    "fetch",
    `${helperReadSource}; return read;`
  )(
    { hostname: page.host, pathname: page.path ?? "/" },
    { querySelectorAll: () => [body] },
    async (url: string) => ({ json: async () => page.fetch?.(url) })
  ) as () => Promise<Result>;
  return read();
}

const inventorApi = (url: string) => {
  if (url.endsWith("/components")) {
    return {
      folders: [
        {
          components: [
            {
              id: "c1",
              name: "ping",
              component_type: "flow",
              current_version_id: "v1",
            },
            {
              id: "c2",
              name: "notes",
              component_type: "text",
              current_version_id: "v2",
            },
          ],
        },
      ],
    };
  }
  return {
    block_list: { blocks: [] },
    trigger: { internal_id: "slash_command", fields: [] },
  };
};

describe("Move to Kite helper", () => {
  it("reads the commands of an inventor.bot dashboard", async () => {
    const seen: string[] = [];
    const result = await run({
      host: "dash.inventor.bot",
      path: "/dash/abc123/components",
      fetch: (url) => (seen.push(url), inventorApi(url)),
    });
    expect(seen[0]).toBe("/api/projects/abc123/components");
    expect(result.message).toBe("Found 1 commands of this bot.");
    // What it copies is what Kite's reader takes.
    const items = parseInventorExport(JSON.stringify(result.data));
    expect(items.map((i) => i.name)).toEqual(["ping"]);
  });

  it("reads the commands of a BotGhost dashboard", async () => {
    const result = await run({
      host: "dashboard.botghost.com",
      state: {
        data: {
          bot: {
            commands: [
              {
                name: "ping",
                actions: [{ type: "plain_text", response: "Pong" }],
              },
            ],
          },
        },
      },
    });
    expect(result.message).toBe(
      "Found 1 commands and 0 custom events of this bot."
    );
    const parsed = parseBotGhostExport(JSON.stringify(result.data));
    expect(parsed.commands.map((c) => c.name)).toEqual(["ping"]);
  });

  it("says where to open it on any other site", async () => {
    const result = await run({ host: "example.com" });
    expect(result.data).toBeUndefined();
    expect(result.message).toContain("BotGhost or inventor.bot");
  });

  it("says so when an inventor.bot page has no bot open", async () => {
    const result = await run({ host: "dash.inventor.bot", path: "/dash" });
    expect(result.data).toBeUndefined();
    expect(result.message).toContain(
      "Open your bot in the inventor.bot dashboard"
    );
  });

  it("is one bookmark that runs as it is", () => {
    expect(helperBookmarklet.startsWith("javascript:")).toBe(true);
    const source = decodeURIComponent(
      helperBookmarklet.slice("javascript:".length)
    );
    expect(source).toBe(helperConsoleSource);
    expect(() => new Function(source)).not.toThrow();
  });
});
