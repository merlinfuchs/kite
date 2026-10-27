import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { FlowTestRequest, FlowTestResponse } from "../types/wire.gen";
import { shortenResult } from "./discord";
import { APICaller, appId, appPath, json } from "./util";

// The API stops a test after 60 seconds.
const testTimeout = 65_000;

// The tool that runs flows as a test. The API reads the app's servers through
// the bot for it, so it needs the discord:read scope.
export function registerFlowTestTool(server: McpServer, { send }: APICaller) {
  server.registerTool(
    "test_flow",
    {
      title: "Test flow",
      description: `Runs a flow as a test without changing anything, to check that it does what it should. The flow runs with a fake trigger: the command with the options, the event of an event listener, a webhook request, or a click on a button or select menu of a message template. Reads from Discord are real, like members and roles. Nothing is sent or changed: messages, role changes and every other change are returned as effects of their block, with warnings when Discord would reject them, like a missing permission or a role above the bot's. Stored variables are read, and changes to them only last for the test. Blocks that call services other than Discord, like Send API Request, AI and integrations, don't run and have no result unless you mock them. Mock blocks to test each branch, like the error branch of an Error Handler or the answers of AI Check and AI Choice. Waits are skipped. Use then to click the buttons, use the select menus and submit the modals the flow showed, which runs those parts of the flow. Returns each run of the flow with the blocks that ran in order, their results, errors and effects. After building or changing a flow, test the branches that matter and fix what fails before telling the user it works.`,
      inputSchema: {
        app_id: appId,
        command_id: z.string().optional(),
        event_listener_id: z.string().optional(),
        message_id: z.string().optional(),
        flow_source_id: z
          .string()
          .optional()
          .describe(
            "With message_id: the flow_source_id of the button or select menu to use."
          ),
        user_id: z
          .string()
          .optional()
          .describe(
            "Who triggers the flow, a member of the server. By default the user of the API token."
          ),
        guild_id: z
          .string()
          .optional()
          .describe(
            "The server, from list_servers. Needed if the bot is in more than one."
          ),
        channel_id: z
          .string()
          .optional()
          .describe(
            "The channel, by default the server's system channel or first text channel."
          ),
        dm: z
          .boolean()
          .optional()
          .describe("Runs a command in a DM with the bot instead."),
        options: z
          .record(z.string(), z.unknown())
          .optional()
          .describe(
            'The arguments of a command by name, like {"user": "123", "amount": 5}. IDs as strings, attachments as the URL of a file.'
          ),
        values: z
          .array(z.string())
          .optional()
          .describe("With message_id: the values picked in the select menu."),
        event: z
          .record(z.string(), z.unknown())
          .optional()
          .describe(
            'Merged into the Discord event of an event listener, or the interaction of a command, in Discord\'s gateway format. Like {"content": "hi"} for message_create, or {"emoji": {"name": "⭐"}, "message_id": "123"} for reactions.'
          ),
        webhook: z
          .object({
            body: z.string(),
            headers: z.record(z.string(), z.string()).optional(),
            query: z.record(z.string(), z.string()).optional(),
          })
          .optional()
          .describe("The request that triggers a webhook event listener."),
        mocks: z
          .record(
            z.string(),
            z.object({
              result: z.unknown().optional(),
              error: z.string().optional(),
            })
          )
          .optional()
          .describe(
            'What blocks do instead of running, by block ID. {"result": ...} is the block\'s result, in the format of its result_schema, like a Discord member object for Get Member, or {"status_code": 200, "body": {...}} for Send API Request. {"error": "..."} makes the block fail. For AI Check the result is "yes", "no" or the probability of yes, for AI Choice the name of an option, and for Verify Webhook "valid" or "invalid".'
          ),
        then: z
          .array(
            z.object({
              click: z.string().optional(),
              select: z.string().optional(),
              values: z.array(z.string()).optional(),
              submit: z.record(z.string(), z.unknown()).optional(),
              user_id: z.string().optional(),
            })
          )
          .max(10)
          .optional()
          .describe(
            'What the user does after the flow ran, in order: {"click": "<label or custom ID of a button>"}, {"select": "<placeholder or custom ID of a select menu>", "values": [...]}, or {"submit": {"<custom ID or label of an input>": "value"}} for the modal the flow showed. user_id makes someone else do it.'
          ),
      },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async (args) => {
      const { app_id, command_id, event_listener_id, message_id } = args;
      const targets = [command_id, event_listener_id, message_id];
      if (targets.filter(Boolean).length !== 1) {
        throw new Error(
          "Pass exactly one of command_id, event_listener_id or message_id."
        );
      }

      const path = command_id
        ? appPath(app_id, "commands", command_id, "test")
        : event_listener_id
        ? appPath(app_id, "event-listeners", event_listener_id, "test")
        : appPath(app_id, "messages", message_id!, "test");

      const res = await send<FlowTestResponse>(
        "POST",
        path,
        {
          flow_source_id: args.flow_source_id,
          user_id: args.user_id,
          guild_id: args.guild_id,
          channel_id: args.channel_id,
          dm: args.dm,
          options: args.options,
          values: args.values,
          event: args.event,
          webhook: args.webhook,
          mocks: args.mocks,
          then: args.then,
        } satisfies FlowTestRequest,
        AbortSignal.timeout(testTimeout)
      );

      return json(shortenResult(withoutEmpty(res), "test"));
    }
  );
}

// Leaves out null and empty values, which Discord objects have many of.
function withoutEmpty(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(withoutEmpty);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .map(([k, v]) => [k, withoutEmpty(v)] as const)
        .filter(
          ([, v]) =>
            v != null &&
            v !== "" &&
            !(Array.isArray(v) && v.length === 0) &&
            !(typeof v === "object" && Object.keys(v).length === 0)
        )
    );
  }
  return value;
}
