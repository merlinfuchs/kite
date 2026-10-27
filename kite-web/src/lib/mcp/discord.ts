import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import {
  discordApiOperations,
  similarDiscordApiOperations,
} from "../flow/discordApi";
import { ApiOperation } from "../integrations/api";
import { matchesSearch } from "../search";
import {
  DiscordAPIRequestRequest,
  DiscordAPIRequestResponse,
} from "../types/wire.gen";
import { APICaller, appId, appPath, json } from "./util";

// Like discordAPIBlockedOutsideFlows in the service, which rejects them.
const blockedOperations = new Set([
  "prune_guild",
  "leave_guild",
  "bulk_delete_messages",
]);

const maxSearchResults = 30;

// The tools that call the Discord API with an app's bot token, for tokens
// with a discord scope. Without write access, only endpoints that read are
// listed.
export function registerDiscordTools(
  server: McpServer,
  { send }: APICaller,
  canWrite: boolean
) {
  const operations = discordApiOperations.filter(
    (o) =>
      !o.interaction &&
      !blockedOperations.has(o.id) &&
      (canWrite || o.method === "GET")
  );
  const byId = new Map(operations.map((o) => [o.id, o]));

  server.registerTool(
    "search_discord_api",
    {
      title: "Search Discord API",
      description: `Finds endpoints of the Discord API that discord_api_request can call, by words of their name or path, like "member role" or "channels messages". Names are operation IDs of Discord's OpenAPI spec, like update_guild for Modify Guild.${
        canWrite
          ? ""
          : " The API token can only read, so only GET endpoints are listed."
      }`,
      inputSchema: { query: z.string() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ query }) => {
      const words = query.replaceAll("_", " ");
      const found = operations.filter((o) =>
        matchesSearch(words, [o.id.replaceAll("_", " "), o.path])
      );
      if (found.length === 0) {
        return json({
          operations: [],
          similar: similarDiscordApiOperations(query.replaceAll(" ", "_"))
            .map((id) => byId.get(id))
            .filter((o) => o !== undefined)
            .map(describe),
        });
      }
      return json({
        operations: found.slice(0, maxSearchResults).map(describe),
        more: found.length > maxSearchResults ? true : undefined,
      });
    }
  );

  server.registerTool(
    "discord_api_request",
    {
      title: "Discord API request",
      description: `Sends a request to an endpoint of the Discord API with the bot token of an app, like Discord's API docs describe, and returns the JSON it returns. Find the endpoint with search_discord_api first. The bot can only do what its roles allow in a server.${
        canWrite
          ? " Changes happen right away and can't be undone, so only make the ones the user asked for. They're logged in the app's logs and the server's audit log."
          : " The API token can only read, so only GET endpoints work."
      }`,
      inputSchema: {
        app_id: appId,
        operation: z.string().describe("The endpoint, like get_guild_member."),
        path_params: z
          .record(z.string(), z.unknown())
          .optional()
          .describe(
            'The parameters in the path, like {"guild_id": "123"}. Give IDs as strings.'
          ),
        query: z
          .record(z.string(), z.unknown())
          .optional()
          .describe("The query parameters, a list for repeated ones."),
        body: z
          .unknown()
          .optional()
          .describe("The JSON body, for endpoints that take one."),
        reason: z
          .string()
          .max(400)
          .optional()
          .describe(
            "Why, shown in the server's audit log for endpoints that write one."
          ),
      },
      annotations: {
        readOnlyHint: !canWrite,
        destructiveHint: canWrite,
        openWorldHint: true,
      },
    },
    async ({ app_id, operation, path_params, query, body, reason }) => {
      const res = await send<DiscordAPIRequestResponse>(
        "POST",
        appPath(app_id, "discord-api"),
        {
          operation,
          path_params: path_params ?? {},
          query: query ?? {},
          body: body ?? null,
          reason: reason ?? "",
        } satisfies DiscordAPIRequestRequest
      );
      return json(res.result);
    }
  );
}

function describe(o: ApiOperation) {
  return {
    id: o.id,
    method: o.method,
    path: o.path,
    path_params: o.path_params.length ? o.path_params : undefined,
    query_params: o.query_params.length ? o.query_params : undefined,
    body_params: o.has_body ? o.body_params ?? "any JSON" : undefined,
  };
}
