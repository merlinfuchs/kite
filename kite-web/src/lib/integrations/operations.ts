import { z } from "zod";
import { ApiOperation } from "./api";
import { integrationApis } from "./apis";
import { handlerOperations } from "./handlerOperations";

// What the service checks when code calls an operation, besides its
// credits. See "Operations for code" in design/integrations.md.
export interface OperationPolicy {
  // What the call acts on, so the service can check it's in the app's server
  // once apps can share a bot: guild, channel, webhook, interaction or user
  // from the path, none for reads of nothing in particular, or unknown when
  // the path doesn't say, like a channel in the body, which needs a rule
  // written by hand first. Not checked yet, apps with their own bot can act
  // anywhere it is.
  target: string;
  // It deletes, removes or bans, judged by its method and path. Calls whose
  // body decides it, like an ER:LC command, aren't marked.
  destructive: boolean;
}

// A request to an integration's API, from its api.json.
export interface RequestOperation extends ApiOperation {
  kind: "request";
  integration: string;
  policy: OperationPolicy;
}

// Go code of the service, see handlerOperations.ts.
export interface HandlerOperation {
  kind: "handler";
  integration: string;
  id: string;
  description: string;
  credits: number;
  // JSON Schemas of the argument and the result.
  params: unknown;
  result: unknown;
  policy: OperationPolicy;
}

export type Operation = RequestOperation | HandlerOperation;

export function operationPolicy(
  integration: string,
  op: ApiOperation
): OperationPolicy {
  const discord = integration === "discord";
  return {
    target: discord ? discordTarget(op) : "none",
    destructive:
      op.method === "DELETE" ||
      (op.method === "PUT" && op.path.includes("/bans/")) ||
      (op.method === "POST" && /bulk-(delete|ban)|\/prune/.test(op.path)),
  };
}

function discordTarget(op: ApiOperation): string {
  if (op.interaction) return "interaction";
  if (op.path.includes("{guild_id}")) return "guild";
  if (op.path.includes("{channel_id}")) return "channel";
  if (op.path.includes("{webhook_id}")) return "webhook";
  if (op.path.startsWith("/users/")) return "user";
  return op.method === "GET" && op.path_params.length === 0
    ? "none"
    : "unknown";
}

// Every operation of every integration, which the service embeds as
// kite-service/pkg/flow/operations.json.
export function allOperations(): Operation[] {
  const requests = Object.entries(integrationApis).flatMap(
    ([integration, api]) =>
      api.operations.map(
        (op): Operation => ({
          kind: "request",
          integration,
          ...op,
          policy: operationPolicy(integration, op),
        })
      )
  );
  const handlers = Object.entries(handlerOperations).flatMap(
    ([integration, ops]) =>
      ops.map(
        (op): Operation => ({
          kind: "handler",
          integration,
          id: op.id,
          description: op.description,
          credits: op.credits,
          params: z.toJSONSchema(op.params),
          result: z.toJSONSchema(op.result),
          policy: { target: "none", destructive: !!op.destructive },
        })
      )
  );
  return [...requests, ...handlers];
}
