import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import type { NextApiRequest, NextApiResponse } from "next";
import { kiteAPI } from "@/lib/mcp/api";
import { createMcpServer } from "@/lib/mcp/server";

// The MCP server for AI clients like Claude Code. Clients authenticate with
// an API token of the user. It keeps no sessions, every request gets its own
// server.
export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== "POST") {
    res.setHeader("Allow", "POST");
    return res.status(405).json(rpcError("Method not allowed, use POST."));
  }

  const token = bearerToken(req.headers.authorization);
  if (!token) {
    return unauthorized(
      res,
      "Send an API token as 'Authorization: Bearer <token>'. Create one in the Kite dashboard under API tokens."
    );
  }

  // Set by nginx in production, which also overwrites one sent by the client.
  const clientIp = req.headers["x-real-ip"];
  const api = kiteAPI(
    token,
    typeof clientIp === "string" ? clientIp : undefined
  );

  // Checked when a client connects, so a wrong token fails there instead of
  // on the first tool call.
  if (isInitialize(req.body)) {
    const check = await api("/v1/apps");
    if (!check.success && check.error.code === "invalid_api_token") {
      return unauthorized(res, check.error.message);
    }
  }

  const server = createMcpServer(api);
  const transport = new StreamableHTTPServerTransport({
    sessionIdGenerator: undefined,
    enableJsonResponse: true,
  });
  res.on("close", () => {
    transport.close();
    server.close();
  });

  await server.connect(transport);
  await transport.handleRequest(req, res, req.body);
}

function isInitialize(body: unknown) {
  const messages = Array.isArray(body) ? body : [body];
  return messages.some((m) => m?.method === "initialize");
}

function unauthorized(res: NextApiResponse, message: string) {
  res.setHeader("WWW-Authenticate", 'Bearer realm="kite"');
  return res.status(401).json(rpcError(message));
}

function bearerToken(header: string | undefined) {
  const [scheme, token] = header?.split(" ") ?? [];
  return scheme?.toLowerCase() === "bearer" && token ? token : undefined;
}

function rpcError(message: string) {
  return { jsonrpc: "2.0", error: { code: -32000, message }, id: null };
}
