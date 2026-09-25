import { FlowContextType } from "./context";
import { getNodeValues } from "./nodes";

export interface NodeCategorySection {
  title: string;
  nodeTypes: string[];
  contextTypes: FlowContextType[] | null;
}

export const nodeCategories: Record<
  "option" | "action" | "control_flow",
  NodeCategorySection[]
> = {
  option: [
    {
      title: "Commands",
      nodeTypes: [
        "option_command_argument",
        "option_command_permissions",
        "option_command_contexts",
      ],
      contextTypes: ["command"],
    },
    {
      title: "Events",
      nodeTypes: ["option_event_filter"],
      contextTypes: ["event_discord"],
    },
    /* {
      title: "Events",
      nodeTypes: ["option_event_filter"],
    }, */
  ],
  action: [
    {
      title: "Responses",
      nodeTypes: [
        "action_response_create",
        "action_response_edit",
        "action_response_delete",
        "action_response_defer",
        "suspend_response_modal",
      ],
      contextTypes: ["command", "component_button", "component_select_menu"],
    },
    {
      title: "Messages",
      nodeTypes: [
        "action_message_create",
        "action_message_edit",
        "action_message_delete",
        "action_message_get",
        "action_private_message_create",
        "action_message_reaction_create",
        "action_message_reaction_delete",
        "action_message_pin",
        "action_message_unpin",
      ],
      contextTypes: null,
    },
    {
      title: "Members",
      nodeTypes: [
        "action_member_ban",
        "action_member_unban",
        "action_member_kick",
        "action_member_timeout",
        "action_member_edit",
        "action_member_get",
      ],
      contextTypes: null,
    },
    {
      title: "Users",
      nodeTypes: ["action_user_get"],
      contextTypes: null,
    },
    {
      title: "Roles",
      nodeTypes: [
        "action_member_role_add",
        "action_member_role_remove",
        "action_role_get",
      ],
      contextTypes: null,
    },

    {
      title: "Servers",
      nodeTypes: ["action_guild_get"],
      contextTypes: null,
    },
    {
      title: "Channels",
      nodeTypes: [
        "action_channel_create",
        "action_channel_edit",
        "action_channel_delete",
        "action_channel_get",
        "action_thread_create",
        "action_thread_member_add",
        "action_thread_member_remove",
      ],
      contextTypes: null,
    },
    {
      title: "Voice",
      nodeTypes: ["action_voice_channel_join", "action_voice_channel_leave"],
      contextTypes: null,
    },
    {
      title: "Bot",
      nodeTypes: ["action_status_set"],
      contextTypes: null,
    },
    {
      title: "Stored Variables",
      nodeTypes: [
        "action_variable_set",
        "action_variable_delete",
        "action_variable_get",
      ],
      contextTypes: null,
    },
    {
      title: "Roblox",
      nodeTypes: ["action_roblox_user_get"],
      contextTypes: null,
    },
    {
      title: "Other Actions",
      nodeTypes: [
        "action_expression_evaluate",
        "action_ai_chat_completion",
        "action_ai_web_search",
        "action_http_request",
        "action_random_generate",
        "action_log",
      ],
      contextTypes: null,
    },
  ],
  control_flow: [
    {
      title: "Conditions",
      nodeTypes: [
        "control_condition_compare",
        "control_condition_user",
        "control_condition_channel",
        "control_condition_role",
      ],
      contextTypes: null,
    },
    {
      title: "Loops",
      nodeTypes: ["control_loop", "control_loop_exit"],
      contextTypes: null,
    },
    {
      title: "Errors",
      nodeTypes: ["control_error_handler"],
      contextTypes: null,
    },
    {
      title: "Others",
      nodeTypes: ["control_sleep"],
      contextTypes: null,
    },
  ],
};

export type NodeCategory = keyof typeof nodeCategories;

function isSectionAvailable(
  section: NodeCategorySection,
  context: FlowContextType
) {
  return !section.contextTypes || section.contextTypes.includes(context);
}

// A block's own contexts take precedence over the sections it's listed in.
// Blocks in neither (e.g. condition items) are always available, as they only
// exist as children of other blocks.
export function isNodeTypeAvailable(type: string, context: FlowContextType) {
  const contexts = getNodeValues(type).contexts;
  if (contexts) return contexts.includes(context);

  const sections = Object.values(nodeCategories)
    .flat()
    .filter((s) => s.nodeTypes.includes(type));
  return (
    sections.length === 0 ||
    sections.some((s) => isSectionAvailable(s, context))
  );
}
