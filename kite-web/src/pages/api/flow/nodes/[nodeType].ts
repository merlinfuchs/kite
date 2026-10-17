import type { NextApiRequest, NextApiResponse } from "next";
import { getNodeInfo, NodeInfo } from "@/lib/flow/nodeInfo";
import env from "@/lib/env/server";

// Only used by the dev server. Static exports get the same data as files from
// scripts/export-node-info.ts.

// CORS middleware function
function corsMiddleware(req: NextApiRequest, res: NextApiResponse) {
  // Allow requests from the docs site
  const allowedOrigins = [env.NEXT_PUBLIC_DOCS_LINK];

  const origin = req.headers.origin;

  if (origin && allowedOrigins.includes(origin)) {
    res.setHeader("Access-Control-Allow-Origin", origin);
  }

  res.setHeader(
    "Access-Control-Allow-Methods",
    "GET, POST, PUT, DELETE, OPTIONS"
  );
  res.setHeader("Access-Control-Allow-Headers", "Content-Type, Authorization");
  res.setHeader("Access-Control-Allow-Credentials", "true");

  // Handle preflight requests
  if (req.method === "OPTIONS") {
    res.status(200).end();
    return true; // Indicates that the request was handled
  }

  return false; // Continue with normal request handling
}

export default function handler(
  req: NextApiRequest,
  res: NextApiResponse<NodeInfo>
) {
  // Handle CORS
  if (corsMiddleware(req, res)) {
    return;
  }

  const { nodeType } = req.query;

  res.status(200).json(getNodeInfo(nodeType as string));
}
