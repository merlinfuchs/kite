import { nodeActionHttpRequestDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const httpRequest: BlockDefinition = {
  type: "action_http_request",
  title: "Send API Request",
  description: "Send an API request to an external server",
  icon: "webhook",
  category: "API Requests",
  credits: 3,
  schema: nodeActionHttpRequestDataSchema,
  inputs: ["http_request_data", "temporary_name", "custom_label"],
  run: { kind: "custom" },
};
