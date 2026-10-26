import {
  componentsGuide,
  placeholdersGuide,
  serializedFlowGuide,
} from "../flow/aiInstructions";

const intro = `This server reads Discord bots made with Kite, a no-code Discord bot builder. Each bot is an app. An app's commands, event listeners and message templates run flows: graphs of blocks with one entry block that starts them, like a slash command, a Discord event, a schedule, a request to a webhook URL, or a click on a button or select menu. Options configure the entry, and actions and controls run after it along the connections.

Start with list_apps, then list the app's commands, event listeners or messages, and get one to see its flow. A message has a flow for each of its buttons and select menus. Use search_blocks and get_block to learn what a block does and what its settings mean, and list_logs to see what happened when a flow ran, like errors. IDs in a block's settings refer to the app's things: variable_id to list_variables, message_template_id to list_messages, and channel and role IDs to list_channels and list_roles. Discord calls servers guilds, so server IDs are guild IDs.

The tools can only read for now. To change an app, tell the user what to change in the Kite editor.`;

const blocksGuide = `get_block describes a block type: title, description, the flow types it can be used in (contexts), its outputs, the blocks it owns (owned_children), data_schema (the JSON Schema of its settings), result_schema, and the integrations it needs, which must be enabled in the app unless it's discord. Result schemas refer to the types listed with them, like #/types/message.`;

// The instructions an MCP client gets when it connects. The sections about
// Kite itself are shared with the flow AI.
export const mcpInstructions = [
  intro,
  serializedFlowGuide,
  blocksGuide,
  componentsGuide,
  placeholdersGuide,
].join("\n\n");
