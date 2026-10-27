import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { Edge, Node } from "@xyflow/react";
import { z } from "zod";
import { blockIntegrations, getBlockDefinition } from "../blocks";
import { buildFlowCatalog } from "../flow/catalog";
import { FlowContextType, flowContextTypes } from "../flow/context";
import { NodeData } from "../flow/dataSchema";
import { serializeFlow } from "../flow/serialize";
import { FlowData } from "../types/flow.gen";
import {
  App,
  Command,
  EventListener,
  LogEntry,
  Message,
  StateGuildChannelListResponse,
  StateGuildListResponse,
  StateGuildRoleListResponse,
  Variable,
} from "../types/wire.gen";
import { KiteAPI } from "./api";
import { mcpInstructions } from "./instructions";

// The tools that read an app. Each call goes to the API with the user's token.
export function createMcpServer(api: KiteAPI) {
  const server = new McpServer(
    { name: "kite", title: "Kite", version: "1.0.0" },
    { instructions: mcpInstructions }
  );

  const read = { readOnlyHint: true, openWorldHint: false };
  const appId = z.string().describe("The ID of the app, from list_apps.");
  const serverId = z
    .string()
    .describe("The ID of the server, from list_servers.");

  const get = async <T>(path: string) => {
    const res = await api<T>(path);
    if (!res.success) {
      throw new Error(`${res.error.message} (${res.error.code})`);
    }
    return res.data;
  };
  const appPath = (id: string, path = "") =>
    `/v1/apps/${encodeURIComponent(id)}${path}`;

  server.registerTool(
    "list_apps",
    {
      title: "List apps",
      description: "Lists the apps of the user, the Discord bots they build.",
      annotations: read,
    },
    async () => {
      const apps = await get<(App | undefined)[]>("/v1/apps");
      return json(
        compact(apps).map((app) => ({
          id: app.id,
          name: app.name,
          description: app.description,
          enabled: app.enabled,
        }))
      );
    }
  );

  server.registerTool(
    "list_commands",
    {
      title: "List commands",
      description:
        "Lists the slash commands of an app, without their flows. Each command and subcommand, like /mod ban, is its own command with its own flow.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const commands = await get<(Command | undefined)[]>(
        appPath(app_id, "/commands")
      );
      return json(
        compact(commands).map((command) => ({
          id: command.id,
          name: command.name,
          description: command.description,
          enabled: command.enabled,
        }))
      );
    }
  );

  server.registerTool(
    "get_command",
    {
      title: "Get command",
      description: "Gets a slash command of an app with its flow.",
      inputSchema: {
        app_id: appId,
        command_id: z.string().describe("The ID of the command."),
      },
      annotations: read,
    },
    async ({ app_id, command_id }) => {
      const command = await get<Command>(
        appPath(app_id, `/commands/${encodeURIComponent(command_id)}`)
      );
      return text(
        JSON.stringify({
          id: command.id,
          name: command.name,
          description: command.description,
          enabled: command.enabled,
        }),
        serialize(command.flow_source, "command")
      );
    }
  );

  server.registerTool(
    "list_event_listeners",
    {
      title: "List event listeners",
      description:
        "Lists the event listeners of an app, without their flows. They run on Discord events, on a schedule or on requests to a webhook URL.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const listeners = await get<(EventListener | undefined)[]>(
        appPath(app_id, "/event-listeners")
      );
      return json(
        compact(listeners).map((listener) => ({
          id: listener.id,
          source: listener.source,
          type: listener.type,
          description: listener.description,
          enabled: listener.enabled,
        }))
      );
    }
  );

  server.registerTool(
    "get_event_listener",
    {
      title: "Get event listener",
      description: "Gets an event listener of an app with its flow.",
      inputSchema: {
        app_id: appId,
        event_listener_id: z.string().describe("The ID of the event listener."),
      },
      annotations: read,
    },
    async ({ app_id, event_listener_id }) => {
      const listener = await get<EventListener>(
        appPath(
          app_id,
          `/event-listeners/${encodeURIComponent(event_listener_id)}`
        )
      );
      return text(
        JSON.stringify({
          id: listener.id,
          source: listener.source,
          type: listener.type,
          description: listener.description,
          enabled: listener.enabled,
          filter: listener.filter,
        }),
        serialize(listener.flow_source, listenerContext(listener.source))
      );
    }
  );

  server.registerTool(
    "list_messages",
    {
      title: "List message templates",
      description:
        "Lists the message templates of an app. Flows send them, and their buttons and select menus run flows of their own.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const messages = await get<(Message | undefined)[]>(
        appPath(app_id, "/messages")
      );
      return json(
        compact(messages).map((message) => ({
          id: message.id,
          name: message.name,
          description: message.description,
        }))
      );
    }
  );

  server.registerTool(
    "get_message",
    {
      title: "Get message template",
      description:
        "Gets a message template of an app: its content, and the flow of each button and select menu by component ID.",
      inputSchema: {
        app_id: appId,
        message_id: z.string().describe("The ID of the message template."),
      },
      annotations: read,
    },
    async ({ app_id, message_id }) => {
      const message = await get<Message>(
        appPath(app_id, `/messages/${encodeURIComponent(message_id)}`)
      );
      return text(
        JSON.stringify({
          id: message.id,
          name: message.name,
          description: message.description,
          data: message.data,
        }),
        ...Object.entries(message.flow_sources).map(
          ([componentId, flow]) =>
            `Flow of component ${componentId}:\n${serialize(
              flow,
              componentContext(flow)
            )}`
        )
      );
    }
  );

  server.registerTool(
    "list_variables",
    {
      title: "List stored variables",
      description:
        "Lists the stored variables of an app, which keep values between runs. Scoped ones keep a value per scope, like per user.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const variables = await get<(Variable | undefined)[]>(
        appPath(app_id, "/variables")
      );
      return json(
        compact(variables).map((variable) => ({
          id: variable.id,
          name: variable.name,
          scoped: variable.scoped,
        }))
      );
    }
  );

  server.registerTool(
    "list_logs",
    {
      title: "List logs",
      description:
        "Lists the newest log entries of an app, like errors of flows, optionally only those of one command, event listener or message template.",
      inputSchema: {
        app_id: appId,
        command_id: z.string().optional(),
        event_listener_id: z.string().optional(),
        message_id: z.string().optional(),
        limit: z
          .number()
          .int()
          .min(1)
          .max(100)
          .optional()
          .describe("Defaults to 20."),
      },
      annotations: read,
    },
    async ({ app_id, command_id, event_listener_id, message_id, limit }) => {
      const query = new URLSearchParams({ limit: String(limit ?? 20) });
      if (command_id) query.set("command_id", command_id);
      if (event_listener_id) query.set("event_id", event_listener_id);
      if (message_id) query.set("message_id", message_id);

      const entries = await get<(LogEntry | undefined)[]>(
        appPath(app_id, `/logs?${query}`)
      );
      return json(
        compact(entries).map((entry) => ({
          level: entry.level,
          message: entry.message,
          command_id: entry.command_id ?? undefined,
          event_listener_id: entry.event_listener_id ?? undefined,
          message_id: entry.message_id ?? undefined,
          block_id: entry.node_id ?? undefined,
          block_type: entry.node_type ?? undefined,
          created_at: entry.created_at,
        }))
      );
    }
  );

  server.registerTool(
    "list_servers",
    {
      title: "List servers",
      description: "Lists the Discord servers the app's bot is in.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const guilds = await get<StateGuildListResponse>(
        appPath(app_id, "/state/guilds")
      );
      return json(
        compact(guilds).map((guild) => ({ id: guild.id, name: guild.name }))
      );
    }
  );

  server.registerTool(
    "list_channels",
    {
      title: "List channels",
      description:
        "Lists the channels of a Discord server the app's bot is in. type is Discord's channel type, like 0 for text and 4 for categories.",
      inputSchema: { app_id: appId, server_id: serverId },
      annotations: read,
    },
    async ({ app_id, server_id }) => {
      const channels = await get<StateGuildChannelListResponse>(
        appPath(
          app_id,
          `/state/guilds/${encodeURIComponent(server_id)}/channels`
        )
      );
      return json(
        compact(channels).map((channel) => ({
          id: channel.id,
          name: channel.name,
          type: channel.type,
        }))
      );
    }
  );

  server.registerTool(
    "list_roles",
    {
      title: "List roles",
      description:
        "Lists the roles of a Discord server the app's bot is in, highest first. Managed roles belong to bots and integrations.",
      inputSchema: { app_id: appId, server_id: serverId },
      annotations: read,
    },
    async ({ app_id, server_id }) => {
      const roles = await get<StateGuildRoleListResponse>(
        appPath(app_id, `/state/guilds/${encodeURIComponent(server_id)}/roles`)
      );
      return json(
        compact(roles)
          .sort((a, b) => b.position - a.position)
          .map((role) => ({
            id: role.id,
            name: role.name,
            managed: role.managed,
          }))
      );
    }
  );

  server.registerTool(
    "search_blocks",
    {
      title: "Search blocks",
      description:
        "Searches the block types flows are built from by words in their type, title or description. Without words it lists all of them by title only.",
      inputSchema: {
        query: z
          .string()
          .optional()
          .describe('Words to search for, like "ban member".'),
        context: z
          .enum(flowContextTypes)
          .optional()
          .describe("Only blocks that can be used in this type of flow."),
      },
      annotations: read,
    },
    async ({ query, context }) => {
      const words = (query ?? "").toLowerCase().split(/\s+/).filter(Boolean);
      const blocks = Object.entries(catalog().nodes)
        .filter(([, block]) => !context || block.contexts.includes(context))
        .filter(([type, block]) => {
          const haystack =
            `${type} ${block.title} ${block.description}`.toLowerCase();
          return words.every((word) => haystack.includes(word));
        });

      if (words.length === 0) {
        return json(
          blocks.map(([type, block]) => ({ type, title: block.title }))
        );
      }
      return json(
        blocks.slice(0, maxSearchResults).map(([type, block]) => ({
          type,
          title: block.title,
          description: block.description,
        }))
      );
    }
  );

  server.registerTool(
    "get_block",
    {
      title: "Get block",
      description:
        "Gets a block type: what it does, where it can be used, its outputs, the JSON Schema of its settings and of its result.",
      inputSchema: {
        type: z.string().describe("The block type, like action_member_ban."),
      },
      annotations: read,
    },
    async ({ type }) => {
      const block = catalog().nodes[type];
      if (!block) {
        throw new Error(
          `Unknown block type ${type}, use search_blocks to find one.`
        );
      }
      const definition = getBlockDefinition(type);
      return json({
        type,
        ...block,
        integrations: definition ? blockIntegrations(definition) : [],
        types: referencedTypes(block),
      });
    }
  );

  return server;
}

const maxSearchResults = 40;

let cachedCatalog: ReturnType<typeof buildFlowCatalog> | undefined;

// Built once, it only changes with a new version of the app.
function catalog() {
  cachedCatalog ??= buildFlowCatalog();
  return cachedCatalog;
}

// The types a block's result schema refers to, like #/types/message, and the
// types those refer to in turn.
function referencedTypes(value: unknown) {
  const types: Record<string, unknown> = {};
  const all = catalog().types as Record<string, unknown>;
  const visit = (v: unknown) => {
    for (const [, name] of JSON.stringify(v).matchAll(/#\/types\/(\w+)/g)) {
      if (name in types || !(name in all)) continue;
      types[name] = all[name];
      visit(all[name]);
    }
  };
  visit(value);
  return types;
}

function serialize(flow: FlowData, context: FlowContextType) {
  return serializeFlow(
    flow.nodes as unknown as Node<NodeData>[],
    flow.edges as unknown as Edge[],
    context
  );
}

function listenerContext(source: string): FlowContextType {
  if (source === "schedule") return "event_schedule";
  if (source === "webhook") return "event_webhook";
  return "event_discord";
}

// Flows of select menus start with their own entry block, all others are
// buttons.
function componentContext(flow: FlowData): FlowContextType {
  return flow.nodes.some((n) => n.type === "entry_component_select_menu")
    ? "component_select_menu"
    : "component_button";
}

function compact<T>(items: (T | undefined)[]) {
  return items.filter((item): item is T => item !== undefined);
}

function json(value: unknown): CallToolResult {
  return text(JSON.stringify(value));
}

function text(...parts: string[]): CallToolResult {
  return { content: [{ type: "text", text: parts.join("\n\n") }] };
}
