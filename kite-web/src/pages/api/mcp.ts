import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import type { NextApiRequest, NextApiResponse } from "next";
import { kiteAPI } from "@/lib/mcp/api";
import { createMcpServer } from "@/lib/mcp/server";
import { APITokenCurrentGetResponse } from "@/lib/types/wire.gen";

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

  // Its scopes decide which tools there are. A wrong token fails when a
  // client connects instead of on the first tool call.
  const current = await api<APITokenCurrentGetResponse>(
    "/v1/users/@me/api-tokens/current"
  );
  if (!current.success) {
    if (current.error.code === "invalid_api_token") {
      return unauthorized(res, current.error.message);
    }
    return res
      .status(502)
      .json(
        rpcError(`Failed to check the API token: ${current.error.message}`)
      );
  }

  const server = createMcpServer(api, current.data.scopes);
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
