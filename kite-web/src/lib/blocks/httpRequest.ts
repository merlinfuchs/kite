import { httpRequestDataSchema } from "../flow/dataSchema";
import { nodeActionHttpRequestResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const httpRequest: BlockDefinition = {
  type: "action_http_request",
  title: "Send API Request",
  description: "Send an API request to an external server",
  icon: "webhook",
  category: "API Requests",
  credits: 3,
  fields: [{ name: "http_request_data", schema: httpRequestDataSchema }],
  result: { schema: nodeActionHttpRequestResultSchema },
  run: { kind: "custom" },
};
