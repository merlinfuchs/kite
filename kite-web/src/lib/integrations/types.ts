import { ZodSchema } from "zod";

// Integration blocks are defined as data: the request they send, the fields
// the user fills in and what they return. The service runs them from the JSON
// that integrations.test.ts writes to kite-service/pkg/flow/integration_blocks.json.
// See design/integrations.md.

export interface Integration {
  id: string;
  name: string;
  blocks: IntegrationBlock[];
}

export type IntegrationFieldType =
  | "snowflake"
  // A list placeholder, or IDs separated by commas or spaces.
  | "snowflake_list"
  | "integer"
  | "boolean"
  | "string";

export interface IntegrationBlockField {
  // Setting in the node's data, like "channel_target" or "max_age". Settings
  // other blocks have too keep their meaning, e.g. channel_target is a channel.
  name: string;
  in: "path" | "query" | "body";
  // Name in the request, if it differs from name.
  target?: string;
  type: IntegrationFieldType;
  label: string;
  description: string;
  required?: boolean;
  // Used when the field is left empty.
  fallback?: "guild" | "channel";
  // Range of an integer, or length of a list.
  min?: number;
  max?: number;
  max_length?: number;
  widget?: "permissions";
}

export interface IntegrationBlock {
  // Node type, namespaced by integration except for Discord.
  type: string;
  integration: string;
  // operationId in the integration's spec.
  operation: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  title: string;
  description: string;
  icon: string;
  // Title of the block explorer section.
  category: string;
  credits: number;
  audit_log_reason?: boolean;
  destructive?: boolean;
  fields: IntegrationBlockField[];
  result?: {
    // Wraps the response like the results of built-in blocks, so e.g.
    // {{result('id').mention}} works.
    thing?: "discord_message" | "discord_role" | "discord_channel";
    list?: boolean;
    schema: ZodSchema;
  };
}
