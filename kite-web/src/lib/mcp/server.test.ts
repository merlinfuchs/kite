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
  version: "abc123",
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

const message = {
  id: "m1",
  name: "Roles",
  description: null,
  data: {
    content: "Pick a role",
    components: [
      {
        type: 1,
        components: [{ type: 6, id: 1, flow_source_id: "menu" }],
      },
    ],
  },
  flow_sources: {},
};

responses["/v1/apps/a1/messages/m1"] = message;

interface Write {
  method: string;
  path: string;
  body: any;
}

// A fake API that answers the paths above and saves writes as they are, and
// a client connected to the server like an MCP client would be.
async function connect(scopes = ["kite:write"]) {
  const paths: string[] = [];
  const writes: Write[] = [];
  const api: KiteAPI = async <T>(path: string, init?: RequestInit) => {
    paths.push(path);
    if (init?.method && init.method !== "GET") {
      const body = init.body ? JSON.parse(init.body as string) : undefined;
      writes.push({ method: init.method, path, body });
      return { success: true, data: { id: "new", ...body } as T };
    }
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
  await createMcpServer(api, scopes).connect(serverTransport);
  const client = new Client({ name: "test", version: "1.0.0" });
  await client.connect(clientTransport);

  const call = async (name: string, args: Record<string, unknown> = {}) => {
    const res = await client.callTool({ name, arguments: args });
    const content = res.content as { type: string; text: string }[];
    return { text: content[0].text, isError: !!res.isError };
  };
  return { client, call, paths, writes };
}

describe("MCP server", () => {
  it("gives clients the instructions and annotated tools", async () => {
    // The longest instructions are the ones for every scope.
    const { client } = await connect(["kite:write", "discord:write"]);
    expect(client.getInstructions()).toContain("list_apps");
    // Claude Code cuts off longer instructions.
    expect(client.getInstructions()!.length).toBeLessThanOrEqual(4000);

    const { tools } = await client.listTools();
    const byName = Object.fromEntries(tools.map((t) => [t.name, t]));
    expect(byName.get_command.annotations?.readOnlyHint).toBe(true);
    expect(byName.create_command.annotations?.readOnlyHint).toBe(false);
    expect(byName.create_command.annotations?.destructiveHint).toBe(false);
    expect(byName.delete_command.annotations?.destructiveHint).toBe(true);
    expect(byName.edit_flow.annotations?.destructiveHint).toBe(true);
  });

  it("has no tools that change apps for read only tokens", async () => {
    const { client } = await connect(["kite:read"]);
    expect(client.getInstructions()).toContain("can only read");

    const { tools } = await client.listTools();
    expect(tools.length).toBeGreaterThan(0);
    for (const tool of tools) {
      expect(tool.annotations?.readOnlyHint, tool.name).toBe(true);
    }
    expect(tools.map((t) => t.name)).not.toContain("get_flow_guide");
  });

  it("has Discord API tools only with a discord scope", async () => {
    const names = async (scopes: string[]) =>
      (await (await connect(scopes)).client.listTools()).tools.map(
        (t) => t.name
      );
    expect(await names(["kite:write"])).not.toContain("discord_api_request");
    expect(await names(["kite:read", "discord:read"])).toContain(
      "discord_api_request"
    );
  });

  it("searches Discord API endpoints that the token can call", async () => {
    const search = async (scopes: string[], query: string) => {
      const { call } = await connect(scopes);
      return JSON.parse((await call("search_discord_api", { query })).text);
    };

    const write = await search(["discord:write"], "member role");
    const ids = write.operations.map((o: { id: string }) => o.id);
    expect(ids).toContain("add_guild_member_role");
    expect(ids).toContain("delete_guild_member_role");

    const read = await search(["discord:read"], "member");
    expect(read.operations.length).toBeGreaterThan(0);
    for (const o of read.operations) expect(o.method).toBe("GET");

    // Blocked ones aren't listed, the API rejects them.
    const prune = await search(["discord:write"], "prune");
    expect(prune.operations.map((o: { id: string }) => o.id)).not.toContain(
      "prune_guild"
    );
  });

  it("sends Discord API requests through the API", async () => {
    const { call, writes } = await connect(["discord:write"]);
    const res = await call("discord_api_request", {
      app_id: "a1",
      operation: "create_message",
      path_params: { channel_id: "123" },
      body: { content: "Hi" },
    });
    expect(res.isError).toBe(false);
    expect(writes).toEqual([
      {
        method: "POST",
        path: "/v1/apps/a1/discord-api",
        body: {
          operation: "create_message",
          path_params: { channel_id: "123" },
          query: {},
          body: { content: "Hi" },
          reason: "",
        },
      },
    ]);
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
    expect(res.text).toContain('"version":"abc123"');
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

  it("gets blocks with the types their results refer to", async () => {
    const { call } = await connect();
    const res = JSON.parse(
      (
        await call("get_blocks", {
          types: ["action_message_create", "action_member_ban", "nope"],
        })
      ).text
    );
    expect(res.blocks.map((b: { type: string }) => b.type)).toEqual([
      "action_message_create",
      "action_member_ban",
    ]);
    expect(res.blocks[0].data_schema).toBeTruthy();
    expect(Object.keys(res.types).length).toBeGreaterThan(0);
    expect(res.unknown_types).toEqual(["nope"]);

    const unknown = await call("get_blocks", { types: ["nope"] });
    expect(unknown.isError).toBe(true);
    expect(unknown.text).toContain("search_blocks");

    const inherited = await call("get_blocks", { types: ["constructor"] });
    expect(inherited.isError).toBe(true);
  });

  it("creates a command with its flow", async () => {
    const { call, writes } = await connect();
    const res = await call("create_command", {
      app_id: "a1",
      name: "hello",
      description: "Says hello",
      edits: [
        {
          op: "add_node",
          ref: "$reply",
          type: "action_response_create",
          data: { message_data: { content: "Hello!" } },
          after: "$entry",
        },
      ],
    });
    expect(res.isError, res.text).toBe(false);

    const [write] = writes;
    expect(write).toMatchObject({
      method: "POST",
      path: "/v1/apps/a1/commands",
    });
    const nodes = write.body.flow_source.nodes;
    expect(nodes.map((n: { type: string }) => n.type)).toEqual([
      "entry_command",
      "action_response_create",
    ]);
    expect(nodes[0].data).toEqual({ name: "hello", description: "Says hello" });
    expect(write.body.flow_source.edges).toHaveLength(1);
    expect(res.text).toContain(`"$reply":"${nodes[1].id}"`);
  });

  it("creates a command disabled", async () => {
    const { call, writes } = await connect();
    const res = await call("create_command", {
      app_id: "a1",
      name: "hello",
      description: "Says hello",
      enabled: false,
    });
    expect(res.isError, res.text).toBe(false);
    expect(writes[0].body.enabled).toBe(false);
  });

  it("explains how to build flows", async () => {
    const { call } = await connect();
    const res = await call("get_flow_guide");
    expect(res.text).toContain("add_node");
    expect(res.text).toContain("arg('name')");
  });

  it("saves nothing when edits cause errors", async () => {
    const { call, writes } = await connect();
    const res = await call("edit_flow", {
      app_id: "a1",
      command_id: "c1",
      edits: [{ op: "add_node", ref: "$x", type: "action_nope", after: "1" }],
    });
    expect(res.isError).toBe(true);
    expect(res.text).toContain("Nothing was saved");
    expect(writes).toHaveLength(0);
  });

  it("edits the flow of a command", async () => {
    const { call, writes } = await connect();
    const res = await call("edit_flow", {
      app_id: "a1",
      command_id: "c1",
      edits: [
        {
          op: "update_node",
          id: "2",
          data: { message_data: { content: "pong!" } },
        },
      ],
    });
    expect(res.isError, res.text).toBe(false);
    expect(writes[0]).toMatchObject({
      method: "PATCH",
      path: "/v1/apps/a1/commands/c1",
    });
    // Fails instead of overwriting a change made since it was read.
    expect(writes[0].body.base_version).toBe(command.version);
    expect(writes[0].body.flow_source.nodes[1].data.message_data.content).toBe(
      "pong!"
    );
  });

  it("needs one flow to edit", async () => {
    const { call } = await connect();
    const res = await call("edit_flow", {
      app_id: "a1",
      command_id: "c1",
      message_id: "m1",
      edits: [],
    });
    expect(res.isError).toBe(true);
    expect(res.text).toContain("exactly one");
  });

  it("builds the flow of a select menu of a message template", async () => {
    const { call, writes } = await connect();
    const res = await call("edit_flow", {
      app_id: "a1",
      message_id: "m1",
      flow_source_id: "menu",
      edits: [
        {
          op: "add_node",
          ref: "$reply",
          type: "action_response_create",
          data: {
            message_data: { content: "You picked <@&{{interaction.value}}>" },
          },
          after: "$entry",
        },
      ],
    });
    expect(res.isError, res.text).toBe(false);
    expect(res.text).toContain("Flow type: component_select_menu");
    expect(writes[0].body.data).toEqual(message.data);
    expect(writes[0].body.flow_sources.menu.nodes).toHaveLength(2);

    const missing = await call("edit_flow", {
      app_id: "a1",
      message_id: "m1",
      flow_source_id: "nope",
      edits: [],
    });
    expect(missing.isError).toBe(true);
  });

  it("creates message templates with flow IDs for their buttons", async () => {
    const { call, writes } = await connect();
    const res = await call("create_message", {
      app_id: "a1",
      name: "Verify",
      data: {
        content: "Click to verify",
        components: [
          {
            type: 1,
            components: [{ type: 2, id: 1, style: 1, label: "Verify" }],
          },
        ],
      },
    });
    expect(res.isError, res.text).toBe(false);
    const button = writes[0].body.data.components[0].components[0];
    expect(button.flow_source_id).toBeTruthy();

    const invalid = await call("create_message", {
      app_id: "a1",
      name: "Bad",
      data: { content: "x".repeat(2001) },
    });
    expect(invalid.isError).toBe(true);
    expect(invalid.text).toContain("content");
  });

  it("creates scheduled event listeners", async () => {
    const { call, writes } = await connect();
    const res = await call("create_event_listener", {
      app_id: "a1",
      event_type: "cron",
      description: "Says good morning",
      cron: "0 9 * * *",
    });
    expect(res.isError, res.text).toBe(false);
    expect(writes[0].body.source).toBe("schedule");
    expect(writes[0].body.flow_source.nodes[0].data).toMatchObject({
      event_type: "cron",
      event_schedule_cron: "0 9 * * *",
    });
  });
});
