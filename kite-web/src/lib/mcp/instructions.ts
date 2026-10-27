import {
  componentsGuide,
  editsGuide,
  placeholdersGuide,
  serializedFlowGuide,
  sharedRules,
} from "../flow/aiInstructions";

const intro = `This server reads and builds Discord bots made with Kite, a no-code Discord bot builder. Each bot is an app. An app's commands, event listeners and message templates run flows: graphs of blocks with one entry block that starts them, like a slash command, a Discord event, a schedule, a request to a webhook URL, or a click on a button or select menu. Options configure the entry, and actions and controls run after it along the connections.

Start with list_apps, then list the app's commands, event listeners or messages, and get one to see its flow. A message has a flow for each of its buttons and select menus. Use search_blocks and get_blocks to learn what a block does and what its settings mean, and list_logs to see what happened when a flow ran, like errors. IDs in a block's settings refer to the app's things: variable_id to list_variables, message_template_id to list_messages, and channel and role IDs to list_channels and list_roles. Discord calls servers guilds, so server IDs are guild IDs.`;

const buildGuide = `To build, create commands and event listeners with their flow in one call, or change existing flows with edit_flow. Call get_flow_guide once before your first edit, it explains edits, placeholders, buttons and the rules for flows. Look up every block type you use with get_blocks first, all in one call, to set its settings exactly as its data_schema says. Changes are saved and live right away, and commands show up in Discord within a minute. If an edit or the flow it results in has problems, nothing is saved and the problems are returned: fix them and send all edits again. The user can restore earlier versions in Kite. Only delete or disable things the user asked you to, and ask before replacing a flow they built.`;

const blocksGuide = `get_blocks describes block types, each with its title, description, the flow types it can be used in (contexts), its outputs, the blocks it owns (owned_children), data_schema (the JSON Schema of its settings), result_schema, and the integrations it needs, which must be enabled in the app unless it's discord. Result schemas refer to the types listed with them, like #/types/message. Settings marked "x-templated" can contain placeholders, and settings marked "x-user-picked" refer to things of the app, like a stored variable: use one from list_variables or create one with create_variable.`;

const messagesGuide = `Message templates are messages a flow sends by their ID in message_template_id, instead of writing the message in the block's message_data. Use them for messages that are reused or have buttons and select menus with longer flows. Each button and select menu of a template has a flow_source_id, set when it's created, and runs the flow of that ID, which edit_flow builds with message_id and flow_source_id. These flows start with an entry_component_button block, also for select menus, and don't use the component_<id> outputs of message blocks.`;

const readOnlyGuide = `The API token of this server can only read apps, not change them. To build or change flows, the user needs a token with write access, which they can create in Kite under API tokens.`;

// The instructions an MCP client gets when it connects, for a token that can
// change apps or not. Claude Code cuts them off after 4096 characters, so how
// to build flows is in flowGuide instead.
export function mcpInstructions(canWrite: boolean) {
  return [
    intro,
    canWrite ? buildGuide : readOnlyGuide,
    serializedFlowGuide,
    blocksGuide,
  ].join("\n\n");
}

// What get_flow_guide returns. The sections about Kite itself are shared with
// the flow AI.
export const flowGuide = [
  editsGuide(false),
  componentsGuide,
  messagesGuide,
  placeholdersGuide,
  `Rules:\n${sharedRules.map((rule) => `- ${rule}`).join("\n")}`,
].join("\n\n");
