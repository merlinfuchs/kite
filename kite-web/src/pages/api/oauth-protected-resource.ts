import type { NextApiRequest, NextApiResponse } from "next";
import { protectedResourceMetadata } from "@/lib/mcp/oauth";

// Served at /.well-known/oauth-protected-resource, see next.config.mjs.
export default function handler(req: NextApiRequest, res: NextApiResponse) {
  // Clients can run in a browser on any site, like the MCP Inspector.
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Cache-Control", "public, max-age=3600");
  res.json(protectedResourceMetadata);
}
