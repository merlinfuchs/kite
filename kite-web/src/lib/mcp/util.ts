import { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { Edge, Node } from "@xyflow/react";
import { z } from "zod";
import { FlowContextType } from "../flow/context";
import { NodeData } from "../flow/dataSchema";
import { serializeFlow } from "../flow/serialize";
import { FlowData } from "../types/flow.gen";
import { Command, EventListener, Message } from "../types/wire.gen";
import { KiteAPI } from "./api";

// Calls the API and throws its errors, which the SDK returns to the client as
// tool errors.
export function apiCaller(api: KiteAPI) {
  const call = async <T>(path: string, init?: RequestInit) => {
    const res = await api<T>(path, init);
    if (!res.success) {
      throw new Error(`${res.error.message} (${res.error.code})`);
    }
    return res.data;
  };

  return {
    get: <T>(path: string) => call<T>(path),
    // Lists of the API can have null items.
    list: async <T>(path: string) =>
      (await call<(T | null | undefined)[]>(path)).filter(
        (item): item is T => item != null
      ),
    send: <T>(method: string, path: string, body?: unknown) =>
      call<T>(path, {
        method,
        body: body === undefined ? undefined : JSON.stringify(body),
        headers: { "Content-Type": "application/json" },
      }),
  };
}

export type APICaller = ReturnType<typeof apiCaller>;

// How many results search tools return at most.
export const maxSearchResults = 40;

export const appId = z.string().describe("The ID of the app, from list_apps.");

// The path of an app's resource, like appPath(appId, "commands", commandId),
// with each part escaped.
export function appPath(appId: string, ...parts: string[]) {
  return ["/v1/apps", ...[appId, ...parts].map(encodeURIComponent)].join("/");
}

export function serialize(flow: FlowData, context: FlowContextType) {
  return serializeFlow(
    flow.nodes as unknown as Node<NodeData>[],
    flow.edges as unknown as Edge[],
    context
  );
}

// What list and get tools show of commands, event listeners and messages.

export function commandInfo(command: Command) {
  return {
    id: command.id,
    name: command.name,
    description: command.description,
    enabled: command.enabled,
  };
}

export function listenerInfo(listener: EventListener) {
  return {
    id: listener.id,
    source: listener.source,
    type: listener.type,
    description: listener.description,
    enabled: listener.enabled,
  };
}

export function messageInfo(message: Message) {
  return {
    id: message.id,
    name: message.name,
    description: message.description,
  };
}

export function json(value: unknown): CallToolResult {
  return text(JSON.stringify(value));
}

// Joins the parts that aren't empty.
export function text(...parts: string[]): CallToolResult {
  return {
    content: [{ type: "text", text: parts.filter(Boolean).join("\n\n") }],
  };
}
