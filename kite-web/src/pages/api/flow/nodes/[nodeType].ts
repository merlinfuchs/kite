import { getNodeValues } from "@/lib/flow/nodes";
import type { NextApiRequest, NextApiResponse } from "next";
import { JsonSchema7Type, zodToJsonSchema } from "zod-to-json-schema";
import env from "@/lib/env/server";

type ResponseData = {
  title: string;
  description: string;
  color: string;
  icon: string;
  dataSchema: JsonSchema7Type | null;
  resultSchema: JsonSchema7Type | null;
  dataFields: string[];
  creditsCost: number | null;
};

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
  res: NextApiResponse<ResponseData>
) {
  // Handle CORS
  if (corsMiddleware(req, res)) {
    return;
  }

  const { nodeType } = req.query;

  const values = getNodeValues(nodeType as string);

  const dataSchema = values.dataSchema
    ? zodToJsonSchema(values.dataSchema, {
        $refStrategy: "none",
      })
    : null;

  const resultSchema = values.resultSchema
    ? zodToJsonSchema(values.resultSchema, {
        $refStrategy: "none",
      })
    : null;

  res.status(200).json({
    title: values.defaultTitle,
    description: values.defaultDescription,
    color: values.color,
    icon: values.icon,
    dataSchema,
    resultSchema,
    dataFields: values.dataFields,
    creditsCost:
      typeof values.creditsCost === "function"
        ? values.creditsCost({})
        : values.creditsCost ?? null,
  });
}
