import { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "../flow/context";
import { NodeData } from "../flow/dataSchema";
import { serializeFlow } from "../flow/serialize";
import { FlowData } from "../types/flow.gen";
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
    send: <T>(method: string, path: string, body?: unknown) =>
      call<T>(path, {
        method,
        body: body === undefined ? undefined : JSON.stringify(body),
        headers: { "Content-Type": "application/json" },
      }),
  };
}

export type APICaller = ReturnType<typeof apiCaller>;

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

export function listenerContext(source: string): FlowContextType {
  if (source === "schedule") return "event_schedule";
  if (source === "webhook") return "event_webhook";
  return "event_discord";
}

export function compact<T>(items: (T | undefined)[]) {
  return items.filter((item): item is T => item !== undefined);
}

export function json(value: unknown): CallToolResult {
  return text(JSON.stringify(value));
}

export function text(...parts: string[]): CallToolResult {
  return { content: [{ type: "text", text: parts.join("\n\n") }] };
}
