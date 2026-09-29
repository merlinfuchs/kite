import { ZodSchema } from "zod";

// Blocks defined as data: their settings, how they run and what they return.
// The service runs them from the JSON that blocks.test.ts writes to
// kite-service/pkg/flow/block_definitions.json. See design/integrations.md.

export type BlockFieldType =
  | "snowflake"
  // A list placeholder, or IDs separated by commas or spaces.
  | "snowflake_list"
  | "integer"
  | "boolean"
  | "string";

export interface BlockField {
  // Setting in the node's data, like "channel_target" or "max_age". Settings
  // other blocks have too keep their meaning, e.g. channel_target is a channel.
  name: string;
  in: "path" | "query" | "body";
  // Name in the request, if it differs from name.
  target?: string;
  type: BlockFieldType;
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

// A request to an integration's API, with the bot's token or the app's
// credential for the integration.
export interface BlockRequest {
  kind: "request";
  // ID of the integration, see ../integrations.
  integration: string;
  // operationId in the integration's spec.
  operation: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
}

export interface BlockDefinition {
  // Node type, namespaced by the integration it acts on except for Discord.
  type: string;
  title: string;
  description: string;
  icon: string;
  // Title of the block explorer section.
  category: string;
  credits: number;
  audit_log_reason?: boolean;
  destructive?: boolean;
  fields: BlockField[];
  run: BlockRequest;
  result?: {
    // Wraps the response like the results of built-in blocks, so e.g.
    // {{result('id').mention}} works.
    thing?: "discord_message" | "discord_role" | "discord_channel";
    list?: boolean;
    schema: ZodSchema;
  };
}
