import type { NextApiRequest, NextApiResponse } from "next";
import { getNodeInfo, NodeInfo } from "@/lib/flow/nodeInfo";
import { isKnownNodeType } from "@/lib/flow/nodes";
import env from "@/lib/env/server";

// Fetched by the block pages of the docs.

// CORS middleware function
function corsMiddleware(req: NextApiRequest, res: NextApiResponse) {
  // Allow requests from the docs site
  const allowedOrigins = [env.NEXT_PUBLIC_DOCS_LINK];

  const origin = req.headers.origin;

  // The response depends on the origin, so caches must not share it.
  res.setHeader("Vary", "Origin");

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
  res: NextApiResponse<NodeInfo | null>
) {
  // Handle CORS
  if (corsMiddleware(req, res)) {
    return;
  }

  const nodeType = String(req.query.nodeType);
  if (!isKnownNodeType(nodeType)) {
    res.status(404).json(null);
    return;
  }

  res.status(200).json(getNodeInfo(nodeType));
}
