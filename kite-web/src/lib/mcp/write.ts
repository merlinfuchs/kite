import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { FlowContextType } from "../flow/context";
import { eventFlowContext } from "../flow/bundleExport";
import { discordEventTypes, eventSourceOfFlow } from "../flow/eventTypes";
import { getWebhookApiUrl } from "../flow/webhook";
import {
  newCommandFlow,
  newComponentFlow,
  newEventListenerFlow,
} from "../flow/newFlows";
import { messageSchema } from "../message/schema";
import { parseMessageData } from "../message/schemaRestore";
import {
  EventTypeScheduleCron,
  EventTypeWebhook,
  FlowData,
} from "../types/flow.gen";
import {
  Command,
  CommandCreateRequest,
  EventListener,
  EventListenerCreateRequest,
  Message,
  MessageCreateRequest,
  MessageUpdateRequest,
  Variable,
  VariableCreateRequest,
} from "../types/wire.gen";
import {
  componentFlowContext,
  describeWarnings,
  editFlow,
  flowEditSchema,
} from "./flows";
import {
  APICaller,
  appId,
  appPath,
  commandInfo,
  json,
  serialize,
  text,
} from "./util";

// The tools that change an app. They save right away like the editor does,
// and every change can be restored from the history of the item in Kite.
export function registerWriteTools(
  server: McpServer,
  { get, send }: APICaller
) {
  const create = {
    readOnlyHint: false,
    destructiveHint: false,
    openWorldHint: false,
  };
  const change = { ...create, destructiveHint: true };

  const edits = z
    .array(flowEditSchema)
    .describe("The edits to make to the flow, applied in order.");

  // Describes a saved flow and what the edits added to it.
  const saved = (
    meta: object,
    flow: FlowData,
    context: FlowContextType,
    res: ReturnType<typeof editFlow>
  ) =>
    text(
      JSON.stringify(meta),
      Object.keys(res.refs).length
        ? `IDs of the added blocks by ref: ${JSON.stringify(res.refs)}`
        : "",
      describeWarnings(res.warnings),
      serialize(flow, context)
    );

  server.registerTool(
    "create_command",
    {
      title: "Create command",
      description:
        "Creates a slash command, optionally with edits that build its flow. A subcommand is its own command named like 'mod ban'. Discord shows new commands within a minute.",
      inputSchema: {
        app_id: appId,
        name: z.string().describe("The name, lowercase, like 'ping'."),
        description: z.string().describe("What it does, shown in Discord."),
        edits: edits.default([]),
      },
      annotations: create,
    },
    async ({ app_id, name, description, edits: flowEdits }) => {
      const res = editFlow(
        newCommandFlow(name, description),
        flowEdits,
        "command"
      );
      const command = await send<Command>("POST", appPath(app_id, "commands"), {
        flow_source: res.flow,
        enabled: true,
      } satisfies CommandCreateRequest);
      return saved(commandInfo(command), command.flow_source, "command", res);
    }
  );

  const eventTypes: [string, ...string[]] = [
    EventTypeScheduleCron,
    EventTypeWebhook,
    ...discordEventTypes.map((t) => t.value),
  ];

  server.registerTool(
    "create_event_listener",
    {
      title: "Create event listener",
      description: `Creates an event listener, optionally with edits that build its flow. It runs on a Discord event, on a schedule (${EventTypeScheduleCron}) or on requests to its webhook URL (${EventTypeWebhook}). The Discord events:\n${discordEventTypes
        .map(
          (t) => `- ${t.value}: ${t.description}.${t.note ? ` ${t.note}` : ""}`
        )
        .join("\n")}`,
      inputSchema: {
        app_id: appId,
        event_type: z.enum(eventTypes),
        description: z.string().describe("What the listener does."),
        cron: z
          .string()
          .optional()
          .describe(
            `For ${EventTypeScheduleCron}: when it runs, as a cron expression in UTC, like '0 9 * * 1' for Mondays at 9:00.`
          ),
        edits: edits.default([]),
      },
      annotations: create,
    },
    async ({ app_id, event_type, description, cron, edits: flowEdits }) => {
      if (cron && event_type !== EventTypeScheduleCron) {
        throw new Error(
          `cron is only for event_type ${EventTypeScheduleCron}.`
        );
      }
      const flow = newEventListenerFlow(event_type, description, cron);
      const source = eventSourceOfFlow(flow);
      const context = eventFlowContext(source);
      const res = editFlow(flow, flowEdits, context);
      const listener = await send<EventListener>(
        "POST",
        appPath(app_id, "event-listeners"),
        {
          source,
          flow_source: res.flow,
          enabled: true,
        } satisfies EventListenerCreateRequest
      );
      return saved(
        {
          id: listener.id,
          webhook_url: getWebhookApiUrl(listener) ?? undefined,
        },
        listener.flow_source,
        context,
        res
      );
    }
  );

  server.registerTool(
    "edit_flow",
    {
      title: "Edit flow",
      description:
        "Edits the flow of a command, an event listener, or a button or select menu of a message template, and saves it. Pass one of command_id, event_listener_id or message_id. If an edit or the resulting flow has problems, nothing is saved and the problems are returned.",
      inputSchema: {
        app_id: appId,
        command_id: z.string().optional(),
        event_listener_id: z.string().optional(),
        message_id: z.string().optional(),
        flow_source_id: z
          .string()
          .optional()
          .describe(
            "With message_id: the flow_source_id of the button or select menu."
          ),
        edits,
      },
      annotations: change,
    },
    async (args) => {
      const { app_id, command_id, event_listener_id, message_id } = args;
      const targets = [command_id, event_listener_id, message_id];
      if (targets.filter(Boolean).length !== 1) {
        throw new Error(
          "Pass exactly one of command_id, event_listener_id or message_id."
        );
      }

      if (command_id) {
        const path = appPath(app_id, "commands", command_id);
        const command = await get<Command>(path);
        const res = editFlow(command.flow_source, args.edits, "command");
        const updated = await send<Command>("PATCH", path, {
          flow_source: res.flow,
        });
        return saved(commandInfo(updated), updated.flow_source, "command", res);
      }

      if (event_listener_id) {
        const path = appPath(app_id, "event-listeners", event_listener_id);
        const listener = await get<EventListener>(path);
        const context = eventFlowContext(listener.source);
        const res = editFlow(listener.flow_source, args.edits, context);
        const updated = await send<EventListener>("PATCH", path, {
          flow_source: res.flow,
        });
        return saved({ id: updated.id }, updated.flow_source, context, res);
      }

      const flowSourceId = args.flow_source_id;
      if (!flowSourceId) {
        throw new Error(
          "Pass the flow_source_id of the button or select menu."
        );
      }
      const path = appPath(app_id, "messages", message_id!);
      const message = await get<Message>(path);
      const context = componentFlowContext(message.data, flowSourceId);
      if (!context) {
        throw new Error(
          `The message has no button or select menu with flow_source_id ${flowSourceId}. Link buttons run no flow.`
        );
      }
      const res = editFlow(
        message.flow_sources[flowSourceId] ?? newComponentFlow(),
        args.edits,
        context
      );
      const updated = await send<Message>("PATCH", path, {
        name: message.name,
        description: message.description,
        data: message.data,
        flow_sources: { ...message.flow_sources, [flowSourceId]: res.flow },
      } satisfies MessageUpdateRequest);
      return saved(
        { id: updated.id, flow_source_id: flowSourceId },
        updated.flow_sources[flowSourceId],
        context,
        res
      );
    }
  );

  const messageData = z
    .record(z.string(), z.unknown())
    .describe(
      "The message like Discord's API takes it: content, embeds, components and flags."
    );

  server.registerTool(
    "create_message",
    {
      title: "Create message template",
      description:
        "Creates a message template, which flows can send. Its buttons and select menus get a flow_source_id, use it with edit_flow to build what they do.",
      inputSchema: {
        app_id: appId,
        name: z.string(),
        description: z.string().optional(),
        data: messageData,
      },
      annotations: create,
    },
    async ({ app_id, name, description, data }) => {
      const message = await send<Message>("POST", appPath(app_id, "messages"), {
        name,
        description: description ?? null,
        data: parseMessage(data),
        flow_sources: {},
      } satisfies MessageCreateRequest);
      return json({ id: message.id, name: message.name, data: message.data });
    }
  );

  server.registerTool(
    "update_message",
    {
      title: "Update message template",
      description:
        "Changes the name, description or content of a message template. data replaces the whole message, so start from get_message. Keep the flow_source_id of existing buttons and select menus, or their flows are deleted.",
      inputSchema: {
        app_id: appId,
        message_id: z.string(),
        name: z.string().optional(),
        description: z.string().optional(),
        data: messageData.optional(),
      },
      annotations: change,
    },
    async ({ app_id, message_id, name, description, data }) => {
      const path = appPath(app_id, "messages", message_id);
      const message = await get<Message>(path);
      const updated = await send<Message>("PATCH", path, {
        name: name ?? message.name,
        description: description ?? message.description,
        data: data ? parseMessage(data) : message.data,
        flow_sources: message.flow_sources,
      } satisfies MessageUpdateRequest);
      return json({ id: updated.id, name: updated.name, data: updated.data });
    }
  );

  server.registerTool(
    "set_enabled",
    {
      title: "Enable or disable",
      description:
        "Enables or disables a command or an event listener. Pass one of command_id or event_listener_id.",
      inputSchema: {
        app_id: appId,
        command_id: z.string().optional(),
        event_listener_id: z.string().optional(),
        enabled: z.boolean(),
      },
      annotations: { ...change, idempotentHint: true },
    },
    async ({ app_id, command_id, event_listener_id, enabled }) => {
      if (!command_id === !event_listener_id) {
        throw new Error("Pass exactly one of command_id or event_listener_id.");
      }
      const path = command_id
        ? appPath(app_id, "commands", command_id, "enabled")
        : appPath(app_id, "event-listeners", event_listener_id!, "enabled");
      await send("PUT", path, { enabled });
      return text(enabled ? "Enabled." : "Disabled.");
    }
  );

  server.registerTool(
    "create_variable",
    {
      title: "Create stored variable",
      description:
        "Creates a stored variable, which keeps a value between runs, for blocks with a variable_id setting.",
      inputSchema: {
        app_id: appId,
        name: z.string(),
        scoped: z
          .boolean()
          .describe(
            "Whether it keeps a value per scope, like per user, which blocks set with their scope setting."
          ),
      },
      annotations: create,
    },
    async ({ app_id, name, scoped }) => {
      const variable = await send<Variable>(
        "POST",
        appPath(app_id, "variables"),
        { name, scoped } satisfies VariableCreateRequest
      );
      return json({
        id: variable.id,
        name: variable.name,
        scoped: variable.scoped,
      });
    }
  );

  const deleteTools = [
    ["delete_command", "command", "commands", "command_id"],
    [
      "delete_event_listener",
      "event listener",
      "event-listeners",
      "event_listener_id",
    ],
    ["delete_message", "message template", "messages", "message_id"],
  ] as const;

  for (const [name, thing, collection, idName] of deleteTools) {
    server.registerTool(
      name,
      {
        title: `Delete ${thing}`,
        description: `Deletes a ${thing}. Only do this when the user asked for it.`,
        inputSchema: { app_id: appId, [idName]: z.string() },
        annotations: change,
      },
      async (args) => {
        const id = (args as Record<string, string>)[idName];
        await send("DELETE", appPath(args.app_id, collection, id));
        return text("Deleted.");
      }
    );
  }
}

// Reads a message like the message editor does, which also gives buttons and
// select menus a flow_source_id, and checks it.
function parseMessage(raw: unknown) {
  let data: ReturnType<typeof parseMessageData>;
  try {
    data = parseMessageData(raw);
  } catch (err) {
    if (err instanceof z.ZodError) {
      throw new Error(`Invalid message:\n${z.prettifyError(err)}`);
    }
    throw err;
  }

  const res = messageSchema.safeParse(data);
  if (!res.success) {
    throw new Error(`Invalid message:\n${z.prettifyError(res.error)}`);
  }
  return data as unknown as MessageCreateRequest["data"];
}
