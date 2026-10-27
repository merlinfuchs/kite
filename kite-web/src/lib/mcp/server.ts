import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { blockIntegrations, getBlockDefinition } from "../blocks";
import { BlockDefinition } from "../blocks/types";
import env from "../env/client";
import { aiModelTiers } from "../flow/aiModels";
import { eventFlowContext } from "../flow/bundleExport";
import { buildFlowCatalog } from "../flow/catalog";
import { flowContextTypes } from "../flow/context";
import { matchesSearch } from "../search";
import {
  App,
  BillingPlan,
  Channel,
  Command,
  EventListener,
  Features,
  Guild,
  LogEntry,
  Message,
  Role,
  UsageByDayEntry,
  UsageByTypeEntry,
  UsageCreditsGetResponse,
  Variable,
} from "../types/wire.gen";
import { hasAccess } from "../apiTokens";
import { KiteAPI } from "./api";
import { componentFlowContext } from "./flows";
import { flowGuide, mcpInstructions } from "./instructions";
// The listing in the MCP registry, published with mcp-publisher. Raise its
// version when publishing changes to it.
import serverInfo from "./server.json";
import {
  apiCaller,
  appPath,
  appId,
  commandInfo,
  json,
  listenerInfo,
  maxSearchResults,
  messageInfo,
  serialize,
  text,
} from "./util";
import { registerDiscordTools } from "./discord";
import { registerTemplateTools } from "./templates";
import { registerWriteTools } from "./write";

// The tools that read and change an app. Each call goes to the API with the
// user's token, and only the tools its scopes allow are listed.
export function createMcpServer(api: KiteAPI, scopes: string[]) {
  const canWrite = hasAccess(scopes, "kite", "write");
  const canReadDiscord = hasAccess(scopes, "discord", "read");
  const canWriteDiscord = hasAccess(scopes, "discord", "write");
  const server = new McpServer(
    { name: "kite", title: serverInfo.title, version: serverInfo.version },
    {
      instructions: mcpInstructions(canWrite, canReadDiscord, canWriteDiscord),
    }
  );

  const read = { readOnlyHint: true, openWorldHint: false };
  const serverId = z
    .string()
    .describe("The ID of the server, from list_servers.");

  const caller = apiCaller(api);
  const { get, list } = caller;

  server.registerTool(
    "list_apps",
    {
      title: "List apps",
      description: "Lists the apps of the user, the Discord bots they build.",
      annotations: read,
    },
    async () => {
      const apps = await list<App>("/v1/apps");
      return json(
        apps.map((app) => ({
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
      const commands = await list<Command>(appPath(app_id, "commands"));
      return json(commands.map(commandInfo));
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
        appPath(app_id, "commands", command_id)
      );
      return text(
        JSON.stringify({ ...commandInfo(command), version: command.version }),
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
      const listeners = await list<EventListener>(
        appPath(app_id, "event-listeners")
      );
      return json(listeners.map(listenerInfo));
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
        appPath(app_id, "event-listeners", event_listener_id)
      );
      return text(
        JSON.stringify({
          ...listenerInfo(listener),
          filter: listener.filter,
          version: listener.version,
        }),
        serialize(listener.flow_source, eventFlowContext(listener.source))
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
      const messages = await list<Message>(appPath(app_id, "messages"));
      return json(messages.map(messageInfo));
    }
  );

  server.registerTool(
    "get_message",
    {
      title: "Get message template",
      description:
        "Gets a message template of an app: its content, and the flow of each button and select menu by its flow_source_id.",
      inputSchema: {
        app_id: appId,
        message_id: z.string().describe("The ID of the message template."),
      },
      annotations: read,
    },
    async ({ app_id, message_id }) => {
      const message = await get<Message>(
        appPath(app_id, "messages", message_id)
      );
      return text(
        JSON.stringify({
          ...messageInfo(message),
          version: message.version,
          data: message.data,
        }),
        // Flows of components that were removed or are link buttons don't run.
        ...Object.entries(message.flow_sources).map(([flowSourceId, flow]) => {
          const context = componentFlowContext(message.data, flowSourceId);
          return context
            ? `Flow of the component with flow_source_id ${flowSourceId}:\n${serialize(
                flow,
                context
              )}`
            : "";
        })
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
      const variables = await list<Variable>(appPath(app_id, "variables"));
      return json(
        variables.map((variable) => ({
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
      if (
        [command_id, event_listener_id, message_id].filter(Boolean).length > 1
      ) {
        throw new Error(
          "Pass at most one of command_id, event_listener_id or message_id."
        );
      }
      const query = new URLSearchParams({ limit: String(limit ?? 20) });
      if (command_id) query.set("command_id", command_id);
      if (event_listener_id) query.set("event_id", event_listener_id);
      if (message_id) query.set("message_id", message_id);

      const entries = await list<LogEntry>(
        `${appPath(app_id, "logs")}?${query}`
      );
      return json(
        entries.map((entry) => ({
          level: entry.level,
          message: entry.message,
          command_id: entry.command_id ?? undefined,
          event_listener_id: entry.event_listener_id ?? undefined,
          message_id: entry.message_id ?? undefined,
          block_id: entry.node_id ?? undefined,
          block_type: entry.node_type ?? undefined,
          error_code: entry.error_code ?? undefined,
          details: entry.details,
          created_at: entry.created_at,
        }))
      );
    }
  );

  server.registerTool(
    "get_usage",
    {
      title: "Get credit usage",
      description:
        "Gets the credits an app used this month, how many its plan has per month, and the credits used each day and by commands, event listeners and messages. Credits reset at the start of each month in UTC, and an app that uses all of them stops running flows until then. get_blocks says what each block costs, and list_plans what other plans include.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const [credits, features, byType, byDay] = await Promise.all([
        get<UsageCreditsGetResponse>(appPath(app_id, "usage", "credits")),
        get<Features>(appPath(app_id, "billing", "features")),
        list<UsageByTypeEntry>(appPath(app_id, "usage", "by-type")),
        list<UsageByDayEntry>(appPath(app_id, "usage", "by-day")),
      ]);
      return json({
        credits_used: credits.credits_used,
        credits_per_month: features.usage_credits_per_month,
        by_flow_type: Object.fromEntries(
          byType.map((entry) => [
            entry.type.replace(/_flow_execution$/, ""),
            entry.credits_used,
          ])
        ),
        // The API lists every day of the month, also the ones to come.
        by_day: byDay
          .filter((entry) => entry.credits_used > 0)
          .map((entry) => ({
            date: entry.date.slice(0, 10),
            credits_used: entry.credits_used,
          })),
      });
    }
  );

  server.registerTool(
    "list_plans",
    {
      title: "List plans",
      description:
        "Lists Kite's plans with their prices in USD and limits, like credits per month, and the page where the user can change the plan of an app. Only bring plans up when the app is near a limit or the user asks, and first suggest ways to need less, like a smaller AI model tier.",
      inputSchema: { app_id: appId },
      annotations: read,
    },
    async ({ app_id }) => {
      const plans = await list<BillingPlan>("/v1/billing/plans");
      return json({
        plans: plans
          .filter((plan) => !plan.hidden)
          .map((plan) => ({
            title: plan.title,
            description: plan.description,
            price_per_month: plan.price,
            price_per_year: plan.yearly_price,
            credits_per_month: plan.feature_usage_credits_per_month,
            max_servers: plan.feature_max_guilds,
            max_commands: plan.feature_max_commands,
            max_event_listeners: plan.feature_max_event_listeners,
            max_scheduled_event_listeners:
              plan.feature_max_scheduled_event_listeners,
            min_schedule_interval_seconds:
              plan.feature_min_schedule_interval_seconds,
            max_webhook_event_listeners:
              plan.feature_max_webhook_event_listeners,
            max_messages: plan.feature_max_messages,
            max_variables: plan.feature_max_variables,
            max_collaborators: plan.feature_max_collaborators,
            max_ai_prompts_per_month: plan.feature_max_ai_prompts_per_month,
          })),
        // Only the user can subscribe, logged in to Kite.
        change_plan_url: `${env.NEXT_PUBLIC_SITE_URL.replace(
          /\/$/,
          ""
        )}/apps/${encodeURIComponent(app_id)}/premium`,
      });
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
      const guilds = await list<Guild>(appPath(app_id, "state", "guilds"));
      return json(guilds.map((guild) => ({ id: guild.id, name: guild.name })));
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
      const channels = await list<Channel>(
        appPath(app_id, "state", "guilds", server_id, "channels")
      );
      return json(
        channels.map((channel) => ({
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
      const roles = await list<Role>(
        appPath(app_id, "state", "guilds", server_id, "roles")
      );
      return json(
        roles
          .sort((a, b) => b.position - a.position)
          .map((role) => ({
            id: role.id,
            name: role.name,
            managed: role.managed,
          }))
      );
    }
  );

  if (canWrite) {
    server.registerTool(
      "get_flow_guide",
      {
        title: "Get flow guide",
        description:
          "Explains how to build flows: the edits, buttons and select menus, message templates, placeholders and rules. Call it once before creating or editing a flow.",
        annotations: read,
      },
      async () => text(flowGuide)
    );
  }

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
      const blocks = Object.entries(catalog().nodes)
        .filter(([, block]) => !context || block.contexts.includes(context))
        .filter(([type, block]) =>
          matchesSearch(query ?? "", [type, block.title, block.description])
        );

      if (!query?.trim()) {
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
    "get_blocks",
    {
      title: "Get blocks",
      description:
        "Gets block types: what each does, where it can be used, its outputs, the JSON Schema of its settings and of its result. Get all the types you need in one call.",
      inputSchema: {
        types: z
          .array(z.string())
          .min(1)
          .max(maxBlocks)
          .describe("The block types, like action_member_ban."),
      },
      annotations: read,
    },
    async ({ types }) => {
      const nodes = catalog().nodes;
      const known = [...new Set(types)].filter((type) =>
        Object.hasOwn(nodes, type)
      );
      const unknown = types.filter((type) => !Object.hasOwn(nodes, type));
      if (!known.length) {
        throw new Error(
          `Unknown block types ${unknown.join(
            ", "
          )}, use search_blocks to find them.`
        );
      }

      const blocks = known.map((type) => {
        const definition = getBlockDefinition(type);
        return {
          type,
          ...nodes[type],
          credits: blockCredits(definition),
          integrations: definition ? blockIntegrations(definition) : [],
        };
      });
      return json({
        blocks,
        // Shared by all blocks, so types they have in common are listed once.
        types: referencedTypes(blocks),
        unknown_types: unknown.length ? unknown : undefined,
      });
    }
  );

  if (canWrite) registerWriteTools(server, caller);
  registerTemplateTools(server, caller, canWrite);
  if (canReadDiscord) registerDiscordTools(server, caller, canWriteDiscord);

  return server;
}

const maxBlocks = 30;

// The credits a block costs each time it runs. Only the cost of AI blocks
// depends on their settings, on the model tier, so it's given for each tier.
function blockCredits(definition: BlockDefinition | undefined) {
  const credits = definition?.credits;
  if (typeof credits !== "function") return credits ?? 0;
  return Object.fromEntries(
    aiModelTiers.map((tier) => [
      tier.value,
      credits({ ai_chat_completion_data: { model: tier.value } }),
    ])
  );
}

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
