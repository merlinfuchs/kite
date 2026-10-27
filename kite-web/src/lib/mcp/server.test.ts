import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";
import { describe, expect, it } from "vitest";
import { KiteAPI } from "./api";
import { createMcpServer } from "./server";

const command = {
  id: "c1",
  name: "ping",
  description: "Replies with pong",
  enabled: true,
  flow_source: {
    nodes: [
      { id: "1", type: "entry_command", data: { name: "ping" } },
      {
        id: "2",
        type: "action_response_create",
        data: { message_data: { content: "pong" } },
      },
    ],
    edges: [{ id: "e1", source: "1", target: "2" }],
  },
};

const responses: Record<string, unknown> = {
  "/v1/apps": [{ id: "a1", name: "Bot", description: null, enabled: true }],
  "/v1/apps/a1/commands/c1": command,
  "/v1/apps/a1/logs?limit=5&command_id=c1": [
    {
      id: 1,
      level: "error",
      message: "Missing permissions",
      command_id: "c1",
      event_listener_id: null,
      message_id: null,
      node_id: "2",
      node_type: "action_response_create",
      created_at: "2026-10-27T10:00:00Z",
    },
  ],
};

// A fake API that answers the paths above, and a client connected to the
// server like an MCP client would be.
async function connect() {
  const paths: string[] = [];
  const api: KiteAPI = async <T>(path: string) => {
    paths.push(path);
    if (path in responses) {
      return { success: true, data: responses[path] as T };
    }
    return {
      success: false,
      error: { code: "unknown_app", message: "App not found", data: {} },
    };
  };

  const [clientTransport, serverTransport] =
    InMemoryTransport.createLinkedPair();
  await createMcpServer(api).connect(serverTransport);
  const client = new Client({ name: "test", version: "1.0.0" });
  await client.connect(clientTransport);

  const call = async (name: string, args: Record<string, unknown> = {}) => {
    const res = await client.callTool({ name, arguments: args });
    const content = res.content as { type: string; text: string }[];
    return { text: content[0].text, isError: !!res.isError };
  };
  return { client, call, paths };
}

describe("MCP server", () => {
  it("gives clients the instructions and read-only tools", async () => {
    const { client } = await connect();
    expect(client.getInstructions()).toContain("list_apps");

    const { tools } = await client.listTools();
    expect(tools.map((t) => t.name)).toContain("get_command");
    for (const tool of tools) {
      expect(tool.annotations?.readOnlyHint, tool.name).toBe(true);
    }
  });

  it("lists apps", async () => {
    const { call } = await connect();
    const res = await call("list_apps");
    expect(JSON.parse(res.text)).toEqual([
      { id: "a1", name: "Bot", description: null, enabled: true },
    ]);
  });

  it("gets a command with its flow as text", async () => {
    const { call } = await connect();
    const res = await call("get_command", { app_id: "a1", command_id: "c1" });
    expect(res.isError).toBe(false);
    expect(res.text).toContain('"name":"ping"');
    expect(res.text).toContain("Flow type: command");
    expect(res.text).toContain("action_response_create");
    expect(res.text).not.toContain("flow_source");
  });

  it("filters logs and names blocks as blocks", async () => {
    const { call, paths } = await connect();
    const res = await call("list_logs", {
      app_id: "a1",
      command_id: "c1",
      limit: 5,
    });
    expect(paths).toContain("/v1/apps/a1/logs?limit=5&command_id=c1");
    expect(JSON.parse(res.text)[0]).toMatchObject({
      level: "error",
      block_id: "2",
      block_type: "action_response_create",
    });
  });

  it("escapes IDs in paths", async () => {
    const { call, paths } = await connect();
    await call("list_commands", { app_id: "../users" });
    expect(paths).toContain("/v1/apps/..%2Fusers/commands");
  });

  it("returns API errors as tool errors", async () => {
    const { call } = await connect();
    const res = await call("list_commands", { app_id: "missing" });
    expect(res.isError).toBe(true);
    expect(res.text).toContain("App not found (unknown_app)");
  });

  it("searches blocks by words and context", async () => {
    const { call } = await connect();
    const res = JSON.parse(
      (await call("search_blocks", { query: "ban member" })).text
    );
    expect(res.map((b: { type: string }) => b.type)).toContain(
      "action_member_ban"
    );

    const all = JSON.parse(
      (await call("search_blocks", { context: "event_schedule" })).text
    );
    const types = all.map((b: { type: string }) => b.type);
    expect(types).not.toContain("option_command_argument");
    expect(all[0]).not.toHaveProperty("description");
  });

  it("gets a block with the types its result refers to", async () => {
    const { call } = await connect();
    const block = JSON.parse(
      (await call("get_block", { type: "action_message_create" })).text
    );
    expect(block.data_schema).toBeTruthy();
    expect(Object.keys(block.types).length).toBeGreaterThan(0);

    const unknown = await call("get_block", { type: "nope" });
    expect(unknown.isError).toBe(true);
    expect(unknown.text).toContain("search_blocks");
  });
});
