import { ZodSchema } from "zod";
import { FlowContextType } from "../flow/context";
import { NodeData } from "../flow/dataSchema";
import { Features } from "../types/wire.gen";

// Every block is defined here: its settings, how it runs, what it returns and
// how it connects to other blocks. The service runs request blocks from the
// JSON that blocks.test.ts writes to kite-service/pkg/flow/block_definitions.json,
// and custom blocks with the Go handler registered under their type. See
// design/integrations.md.

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

// Runs the Go handler registered under the block's type.
export interface BlockCustomRun {
  kind: "custom";
}

// Canvas components, see components.ts. Blocks without one use "action".
export type BlockComponent =
  | "action"
  | "action_message"
  | "option"
  | "option_command_argument"
  | "entry_command"
  | "entry_event"
  | "entry_component_button"
  | "condition_compare"
  | "condition_user"
  | "condition_channel"
  | "condition_role"
  | "condition_item"
  | "control_loop"
  | "control_loop_each"
  | "control_loop_end"
  | "control_loop_exit"
  | "control_sleep"
  | "control_error_handler"
  | "suspend";

export interface BlockDefinition {
  // Node type, namespaced by the integration it acts on except for Discord.
  type: string;
  title: string;
  description: string;
  icon: string;
  // Defaults to the color of the block's kind, e.g. actions are blue.
  color?: string;
  // Title of the block explorer section. Blocks without one are only created
  // with other blocks, like the branches of a condition.
  category?: string;
  // Flow types the block can be used in, otherwise those of its section.
  contexts?: FlowContextType[];
  // Output handles, defaults to ["default"].
  outputs?: string[];
  // Blocks created together with this one and connected to it with fixed
  // edges, like the else branch of a condition.
  owns?: string[];
  // Can't be deleted or duplicated, like the entry block.
  fixed?: boolean;
  component?: BlockComponent;
  // Integrations the block needs besides the one its request goes to.
  requires?: string[];
  // The block fails when the app doesn't have this feature.
  premium_feature?: keyof Features;
  credits?: number | ((data: NodeData) => number);
  // Settings either as fields, whose schema and inputs are generated, or as a
  // schema with the names of the editor inputs (widgets) that edit it. A
  // schema can be a function, for schemas that depend on other blocks.
  fields?: BlockField[];
  schema?: ZodSchema | (() => ZodSchema);
  inputs?: string[];
  audit_log_reason?: boolean;
  run: BlockRequest | BlockCustomRun;
  result?: {
    // Wraps the response of a request like the results of built-in blocks,
    // so e.g. {{result('id').mention}} works.
    thing?: "discord_message" | "discord_role" | "discord_channel";
    list?: boolean;
    schema: ZodSchema;
  };
}

// A block that sends a request, like those defined with fields.
export type RequestBlockDefinition = BlockDefinition & {
  run: BlockRequest;
  fields: BlockField[];
  // The service reads it for requests, so it can't depend on the settings.
  credits: number;
};
