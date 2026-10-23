import { Edge, Node } from "@xyflow/react";
import { getLayoutedElements } from "./layout";
import { getEdgeId, getNodeId } from "./nodes";
import { NodeData } from "./dataSchema";
import {
  BrainCircuitIcon,
  GavelIcon,
  LucideIcon,
  MailIcon,
  SirenIcon,
  UserRoundPlusIcon,
} from "lucide-react";

export type TemplateCategory =
  | "moderation"
  | "ai"
  | "community"
  | "support"
  | "games";

export const templateCategories: { value: TemplateCategory; label: string }[] =
  [
    { value: "moderation", label: "Moderation" },
    { value: "ai", label: "AI" },
    { value: "community", label: "Community" },
    { value: "support", label: "Support" },
    { value: "games", label: "Games" },
  ];

export type Template = {
  // id is used in links and analytics, so it must never change.
  id: string;
  // version goes up whenever the flows change. Installed copies don't update.
  version: number;
  name: string;
  description: string;
  icon: LucideIcon;
  category: TemplateCategory;
  // setup are the steps outside of the template that it needs to work.
  setup?: string[];
  inputs: {
    key: string;
    label: string;
    description: string;
    type: "text" | "textarea";
    required: boolean;
  }[];
  commands: {
    name: string;
    description: string;
    flowSource(inputs: Record<string, any>): {
      nodes: Omit<Node<NodeData>, "position">[];
      edges: Edge[];
    };
  }[];
  eventListeners: {
    source: string;
    type: string;
    description: string;
    flowSource(inputs: Record<string, any>): {
      nodes: Omit<Node<NodeData>, "position">[];
      edges: Edge[];
    };
  }[];
};

export function getTemplates() {
  return [
    getModerationTemplate(),
    getAITemplate(),
    getWelcomerTemplate(),
    getModmailTemplate(),
    getERLCTemplate(),
  ];
}

// templateItemKey identifies a command or event listener within its template,
// so a later version of the template can be matched to what an app installed.
export function templateItemKey(
  item: Template["commands"][number] | Template["eventListeners"][number]
) {
  return "type" in item ? item.type : item.name;
}

export function getTemplate(id: string) {
  return getTemplates().find((t) => t.id === id);
}

export function prepareTemplateFlow(flow: {
  nodes: Omit<Node<NodeData>, "position">[];
  edges: Edge[];
}) {
  return getLayoutedElements(
    flow.nodes.map((node) => ({
      ...node,
      position: { x: 0, y: 0 },
    })),
    flow.edges
  );
}

export function getModerationTemplate(): Template {
  const moderationBanEntryNodeId = getNodeId("entry_command");
  const moderationBanOptionUserIdNodeId = getNodeId("option_command_argument");
  const moderationBanOptionPermissionsNodeId = getNodeId(
    "option_command_permissions"
  );
  const moderationBanOptionReasonNodeId = getNodeId("option_command_argument");
  const moderationBanActionMemberBanNodeId = getNodeId("action_member_ban");
  const moderationBanActionResponseNodeId = getNodeId("action_response_create");

  const moderationUnbanEntryNodeId = getNodeId("entry_command");
  const moderationUnbanOptionUserIdNodeId = getNodeId(
    "option_command_argument"
  );
  const moderationUnbanOptionPermissionsNodeId = getNodeId(
    "option_command_permissions"
  );
  const moderationUnbanOptionReasonNodeId = getNodeId(
    "option_command_argument"
  );
  const moderationUnbanActionMemberUnbanNodeId = getNodeId(
    "action_member_unban"
  );
  const moderationUnbanActionResponseNodeId = getNodeId(
    "action_response_create"
  );

  const moderationKickEntryNodeId = getNodeId("entry_command");
  const moderationKickOptionUserIdNodeId = getNodeId("option_command_argument");
  const moderationKickOptionPermissionsNodeId = getNodeId(
    "option_command_permissions"
  );
  const moderationKickOptionReasonNodeId = getNodeId("option_command_argument");
  const moderationKickActionMemberKickNodeId = getNodeId("action_member_kick");
  const moderationKickActionResponseNodeId = getNodeId(
    "action_response_create"
  );

  const moderationMuteEntryNodeId = getNodeId("entry_command");
  const moderationMuteOptionUserIdNodeId = getNodeId("option_command_argument");
  const moderationMuteOptionPermissionsNodeId = getNodeId(
    "option_command_permissions"
  );
  const moderationMuteOptionDurationNodeId = getNodeId(
    "option_command_argument"
  );
  const moderationMuteOptionReasonNodeId = getNodeId("option_command_argument");
  const moderationMuteActionMemberTimeoutNodeId = getNodeId(
    "action_member_timeout"
  );
  const moderationMuteActionResponseNodeId = getNodeId(
    "action_response_create"
  );
  return {
    id: "moderation",
    version: 1,
    name: "Moderation",
    description:
      "Ban, unban, kick and mute commands, limited to members with the matching permission.",
    icon: GavelIcon,
    category: "moderation",
    inputs: [],
    commands: [
      {
        name: "ban",
        description: "Ban a user from the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: moderationBanEntryNodeId,
              type: "entry_command",
              data: {
                name: "ban",
                description: "Ban a user from the server.",
              },
            },
            {
              id: moderationBanOptionUserIdNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to ban.",
                command_argument_type: "user",
                command_argument_required: true,
              },
            },
            {
              id: moderationBanOptionReasonNodeId,
              type: "option_command_argument",
              data: {
                name: "reason",
                description: "The reason for the ban.",
                command_argument_type: "string",
                command_argument_required: false,
              },
            },
            {
              id: moderationBanOptionPermissionsNodeId,
              type: "option_command_permissions",
              data: {
                command_permissions: "4",
              },
            },
            {
              id: moderationBanActionMemberBanNodeId,
              type: "action_member_ban",
              data: {
                user_target: "{{interaction.command.args.user}}",
                audit_log_reason: "{{interaction.command.args.reason}}",
                member_ban_delete_message_duration_seconds: "3600",
              },
            },
            {
              id: moderationBanActionResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "The user {{interaction.command.args.user.mention}} has been banned.",
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: moderationBanOptionUserIdNodeId,
              target: moderationBanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationBanOptionPermissionsNodeId,
              target: moderationBanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationBanOptionReasonNodeId,
              target: moderationBanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationBanEntryNodeId,
              target: moderationBanActionMemberBanNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationBanActionMemberBanNodeId,
              target: moderationBanActionResponseNodeId,
            },
          ],
        }),
      },
      {
        name: "unban",
        description: "Unban a user from the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: moderationUnbanEntryNodeId,
              type: "entry_command",
              data: {
                name: "unban",
                description: "Unban a user from the server.",
              },
            },
            {
              id: moderationUnbanOptionUserIdNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to unban.",
                command_argument_type: "user",
                command_argument_required: true,
              },
            },
            {
              id: moderationUnbanOptionReasonNodeId,
              type: "option_command_argument",
              data: {
                name: "reason",
                description: "The reason for the unban.",
                command_argument_type: "string",
                command_argument_required: false,
              },
            },
            {
              id: moderationUnbanOptionPermissionsNodeId,
              type: "option_command_permissions",
              data: {
                command_permissions: "4",
              },
            },
            {
              id: moderationUnbanActionMemberUnbanNodeId,
              type: "action_member_unban",
              data: {
                user_target: "{{interaction.command.args.user}}",
                audit_log_reason: "{{interaction.command.args.reason}}",
              },
            },
            {
              id: moderationUnbanActionResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "The user {{interaction.command.args.user.mention}} has been unbanned.",
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: moderationUnbanOptionUserIdNodeId,
              target: moderationUnbanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationUnbanOptionReasonNodeId,
              target: moderationUnbanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationUnbanOptionPermissionsNodeId,
              target: moderationUnbanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationUnbanEntryNodeId,
              target: moderationUnbanActionMemberUnbanNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationUnbanActionMemberUnbanNodeId,
              target: moderationUnbanActionResponseNodeId,
            },
          ],
        }),
      },
      {
        name: "kick",
        description: "Kick a user from the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: moderationKickEntryNodeId,
              type: "entry_command",
              data: {
                name: "kick",
                description: "Kick a user from the server.",
              },
            },
            {
              id: moderationKickOptionUserIdNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to kick.",
                command_argument_type: "user",
                command_argument_required: true,
              },
            },
            {
              id: moderationKickOptionReasonNodeId,
              type: "option_command_argument",
              data: {
                name: "reason",
                description: "The reason for the kick.",
                command_argument_type: "string",
                command_argument_required: false,
              },
            },
            {
              id: moderationKickOptionPermissionsNodeId,
              type: "option_command_permissions",
              data: {
                command_permissions: "2",
              },
            },
            {
              id: moderationKickActionMemberKickNodeId,
              type: "action_member_kick",
              data: {
                user_target: "{{interaction.command.args.user}}",
                audit_log_reason: "{{interaction.command.args.reason}}",
              },
            },
            {
              id: moderationKickActionResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "The user {{interaction.command.args.user.mention}} has been kicked.",
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: moderationKickOptionUserIdNodeId,
              target: moderationKickEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationKickOptionReasonNodeId,
              target: moderationKickEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationKickOptionPermissionsNodeId,
              target: moderationKickEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationKickEntryNodeId,
              target: moderationKickActionMemberKickNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationKickActionMemberKickNodeId,
              target: moderationKickActionResponseNodeId,
            },
          ],
        }),
      },
      {
        name: "mute",
        description: "Mute a user in the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: moderationMuteEntryNodeId,
              type: "entry_command",
              data: {
                name: "mute",
                description: "Mute a user in the server.",
              },
            },
            {
              id: moderationMuteOptionUserIdNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to mute.",
                command_argument_type: "user",
                command_argument_required: true,
              },
            },
            {
              id: moderationMuteOptionDurationNodeId,
              type: "option_command_argument",
              data: {
                name: "duration",
                description: "The number of seconds to mute the user for.",
                command_argument_type: "number",
                command_argument_required: true,
              },
            },
            {
              id: moderationMuteOptionReasonNodeId,
              type: "option_command_argument",
              data: {
                name: "reason",
                description: "The reason for the mute.",
                command_argument_type: "string",
                command_argument_required: false,
              },
            },
            {
              id: moderationMuteOptionPermissionsNodeId,
              type: "option_command_permissions",
              data: {
                command_permissions: "1099511627776",
              },
            },
            {
              id: moderationMuteActionMemberTimeoutNodeId,
              type: "action_member_timeout",
              data: {
                user_target: "{{interaction.command.args.user}}",
                member_timeout_duration_seconds:
                  "{{interaction.command.args.duration}}",
                audit_log_reason: "{{interaction.command.args.reason}}",
              },
            },
            {
              id: moderationMuteActionResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "The user {{interaction.command.args.user.mention}} has been muted for `{{interaction.command.args.duration}}` seconds.",
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: moderationMuteOptionUserIdNodeId,
              target: moderationMuteEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationMuteOptionReasonNodeId,
              target: moderationMuteEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationMuteOptionDurationNodeId,
              target: moderationMuteEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationMuteOptionPermissionsNodeId,
              target: moderationMuteEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationMuteEntryNodeId,
              target: moderationMuteActionMemberTimeoutNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationMuteActionMemberTimeoutNodeId,
              target: moderationMuteActionResponseNodeId,
            },
          ],
        }),
      },
    ],
    eventListeners: [],
  };
}

export function getAITemplate(): Template {
  const aiAskCommandEntryNodeId = getNodeId("entry_command");
  const aiAskCommandOptionQuestionNodeId = getNodeId("option_command_argument");
  const aiAskCommandActionAiChatCompletionNodeId = getNodeId(
    "action_ai_chat_completion"
  );
  const aiAskCommandActionResponseCreateNodeId = getNodeId(
    "action_response_create"
  );

  const aiAskEventEntryNodeId = getNodeId("entry_event");
  const aiAskEventFilterNodeId = getNodeId("option_event_filter");
  const aiAskEventActionAiChatCompletionNodeId = getNodeId(
    "action_ai_chat_completion"
  );
  const aiAskEventActionTypingTriggerNodeId = getNodeId(
    "action_typing_trigger"
  );
  const aiAskEventActionMessageCreateNodeId = getNodeId(
    "action_message_create"
  );

  return {
    id: "ask-ai",
    version: 1,
    name: "Ask AI",
    description: "Members ask AI questions with /ask or by mentioning the bot.",
    icon: BrainCircuitIcon,
    category: "ai",
    inputs: [
      {
        key: "system_prompt",
        label: "Personality",
        description:
          "Give the AI a personality. Tell it how to respond to questions.",
        type: "textarea",
        required: false,
      },
    ],
    commands: [
      {
        name: "ask",
        description: "Ask a question to the AI.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: aiAskCommandEntryNodeId,
              type: "entry_command",
              data: {
                name: "ask",
                description: "Ask a question to the AI.",
              },
            },
            {
              id: aiAskCommandOptionQuestionNodeId,
              type: "option_command_argument",
              data: {
                name: "question",
                description: "The question to ask the AI.",
                command_argument_type: "string",
                command_argument_required: true,
              },
            },
            {
              id: aiAskCommandActionAiChatCompletionNodeId,
              type: "action_ai_chat_completion",
              data: {
                ai_chat_completion_data: {
                  system_prompt: inputs.system_prompt ?? undefined,
                  prompt: "{{interaction.command.args.question}}",
                },
              },
            },
            {
              id: aiAskCommandActionResponseCreateNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content: `{{nodes.${aiAskCommandActionAiChatCompletionNodeId}.result}}`,
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: aiAskCommandOptionQuestionNodeId,
              target: aiAskCommandEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskCommandEntryNodeId,
              target: aiAskCommandActionAiChatCompletionNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskCommandActionAiChatCompletionNodeId,
              target: aiAskCommandActionResponseCreateNodeId,
            },
          ],
        }),
      },
    ],
    eventListeners: [
      {
        source: "discord",
        type: "message_create",
        description: "Ask a question to the AI by pinging the bot.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: aiAskEventEntryNodeId,
              type: "entry_event",
              data: {
                event_type: "message_create",
                description: "Ask a question to the AI by pinging the bot.",
              },
            },
            // A filter is checked before the flow starts, so messages that
            // don't ping the bot don't run the flow.
            {
              id: aiAskEventFilterNodeId,
              type: "option_event_filter",
              data: {
                event_filter_target: "message_mentions_bot",
                event_filter_mode: "equal",
              },
            },
            {
              id: aiAskEventActionAiChatCompletionNodeId,
              type: "action_ai_chat_completion",
              data: {
                ai_chat_completion_data: {
                  system_prompt: inputs.system_prompt ?? undefined,
                  prompt: "{{event.message.content}}",
                },
              },
            },
            {
              id: aiAskEventActionTypingTriggerNodeId,
              type: "action_typing_trigger",
              data: {
                channel_target: "{{event.channel.id}}",
              },
            },
            {
              id: aiAskEventActionMessageCreateNodeId,
              type: "action_message_create",
              data: {
                channel_target: "{{event.channel.id}}",
                reply_target: "{{event.message.id}}",
                message_data: {
                  content: `{{nodes.${aiAskEventActionAiChatCompletionNodeId}.result}}`,
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: aiAskEventFilterNodeId,
              target: aiAskEventEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventEntryNodeId,
              target: aiAskEventActionAiChatCompletionNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventActionAiChatCompletionNodeId,
              target: aiAskEventActionTypingTriggerNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventActionTypingTriggerNodeId,
              target: aiAskEventActionMessageCreateNodeId,
            },
          ],
        }),
      },
    ],
  };
}

export function getWelcomerTemplate(): Template {
  const welcomerEntryNodeId = getNodeId("entry_event");
  const welcomerActionMessageCreateNodeId = getNodeId("action_message_create");

  return {
    id: "welcomer",
    version: 1,
    name: "Welcomer",
    description:
      "Greets every new member with a message in a channel you choose.",
    icon: UserRoundPlusIcon,
    category: "community",
    inputs: [
      {
        key: "channel_id",
        label: "Channel ID",
        description: "The channel to send the welcome messages to.",
        type: "text",
        required: true,
      },
    ],
    commands: [],
    eventListeners: [
      {
        source: "discord",
        type: "guild_member_add",
        description: "Welcome a new user to the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: welcomerEntryNodeId,
              type: "entry_event",
              data: {
                event_type: "guild_member_add",
                description: "Welcome a new user to the server.",
              },
            },
            {
              id: welcomerActionMessageCreateNodeId,
              type: "action_message_create",
              data: {
                channel_target: inputs.channel_id,
                message_data: {
                  content: "Welcome {{event.user.mention}} to the server!",
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: welcomerEntryNodeId,
              target: welcomerActionMessageCreateNodeId,
            },
          ],
        }),
      },
    ],
  };
}

export function getModmailTemplate(): Template {
  const modmailEventEntryNodeId = getNodeId("entry_event");
  const modmailEventForwardNodeId = getNodeId("action_message_create");
  const modmailEventConfirmNodeId = getNodeId("action_private_message_create");

  const modmailReplyEntryNodeId = getNodeId("entry_command");
  const modmailReplyOptionUserNodeId = getNodeId("option_command_argument");
  const modmailReplyOptionMessageNodeId = getNodeId("option_command_argument");
  const modmailReplyOptionPermissionsNodeId = getNodeId(
    "option_command_permissions"
  );
  const modmailReplyActionSendNodeId = getNodeId(
    "action_private_message_create"
  );
  const modmailReplyActionResponseNodeId = getNodeId("action_response_create");
  const modmailReplyErrorHandlerNodeId = getNodeId("control_error_handler");
  const modmailReplyErrorResponseNodeId = getNodeId("action_response_create");

  return {
    id: "modmail",
    version: 1,
    name: "Modmail",
    description:
      "Users DM the bot to reach your staff, and staff answer with /reply.",
    icon: MailIcon,
    category: "support",
    inputs: [
      {
        key: "channel_id",
        label: "Staff Channel ID",
        description:
          "The channel the bot posts the direct messages it receives in. Only staff should be able to see it.",
        type: "text",
        required: true,
      },
    ],
    commands: [
      {
        name: "reply",
        description: "Reply to a modmail message.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: modmailReplyEntryNodeId,
              type: "entry_command",
              data: {
                name: "reply",
                description: "Reply to a modmail message.",
              },
            },
            {
              id: modmailReplyOptionUserNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to reply to.",
                command_argument_type: "user",
                command_argument_required: true,
              },
            },
            {
              id: modmailReplyOptionMessageNodeId,
              type: "option_command_argument",
              data: {
                name: "message",
                description: "Your reply.",
                command_argument_type: "string",
                command_argument_required: true,
              },
            },
            {
              id: modmailReplyOptionPermissionsNodeId,
              type: "option_command_permissions",
              data: {
                // Manage Messages
                command_permissions: "8192",
              },
            },
            {
              id: modmailReplyActionSendNodeId,
              type: "action_private_message_create",
              data: {
                user_target: "{{interaction.command.args.user}}",
                message_data: {
                  content:
                    "**Reply from the staff of {{interaction.guild.name}}:**\n{{interaction.command.args.message}}",
                },
              },
            },
            {
              id: modmailReplyActionResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "Sent your reply to {{interaction.command.args.user.mention}}:\n>>> {{interaction.command.args.message}}",
                },
              },
            },
            {
              id: modmailReplyErrorHandlerNodeId,
              type: "control_error_handler",
              data: {},
            },
            {
              id: modmailReplyErrorResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content:
                    "Couldn't send your reply to {{interaction.command.args.user.mention}}. They may have DMs turned off or don't share a server with the bot.",
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: modmailReplyOptionUserNodeId,
              target: modmailReplyEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyOptionMessageNodeId,
              target: modmailReplyEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyOptionPermissionsNodeId,
              target: modmailReplyEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyEntryNodeId,
              target: modmailReplyErrorHandlerNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyErrorHandlerNodeId,
              sourceHandle: "default",
              target: modmailReplyActionSendNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyActionSendNodeId,
              target: modmailReplyActionResponseNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailReplyErrorHandlerNodeId,
              sourceHandle: "error",
              target: modmailReplyErrorResponseNodeId,
            },
          ],
        }),
      },
    ],
    eventListeners: [
      {
        source: "discord",
        type: "direct_message_create",
        description: "Forward direct messages to the staff channel.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: modmailEventEntryNodeId,
              type: "entry_event",
              data: {
                event_type: "direct_message_create",
                description: "Forward direct messages to the staff channel.",
              },
            },
            {
              id: modmailEventForwardNodeId,
              type: "action_message_create",
              data: {
                channel_target: inputs.channel_id,
                message_data: {
                  content:
                    "📬 **New modmail from {{event.user.mention}}** (`{{event.user.id}}`)\n>>> {{event.message.content}}\n\nAnswer with `/reply`.",
                  allowed_mentions: { parse: [] },
                },
              },
            },
            {
              id: modmailEventConfirmNodeId,
              type: "action_private_message_create",
              data: {
                user_target: "{{event.user.id}}",
                message_data: {
                  content:
                    "Thanks, your message was sent to the staff team. You'll get a reply here.",
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: modmailEventEntryNodeId,
              target: modmailEventForwardNodeId,
            },
            {
              id: getEdgeId(),
              source: modmailEventForwardNodeId,
              target: modmailEventConfirmNodeId,
            },
          ],
        }),
      },
    ],
  };
}

export function getERLCTemplate(): Template {
  // Each command runs its blocks under an error handler, so a server key that
  // isn't set up or an offline server gets a reply instead of a generic error.
  function command(
    name: string,
    description: string,
    options: Omit<Node<NodeData>, "position">[],
    actions: Omit<Node<NodeData>, "position">[],
    errorMessage: (errorNodeId: string) => string
  ) {
    const entryNodeId = getNodeId("entry_command");
    const errorHandlerNodeId = getNodeId("control_error_handler");
    const errorResponseNodeId = getNodeId("action_response_create");

    return {
      name,
      description,
      flowSource: () => ({
        nodes: [
          {
            id: entryNodeId,
            type: "entry_command",
            data: { name, description },
          },
          ...options,
          {
            id: errorHandlerNodeId,
            type: "control_error_handler",
            data: {},
          },
          ...actions,
          {
            id: errorResponseNodeId,
            type: "action_response_create",
            data: {
              message_data: { content: errorMessage(errorHandlerNodeId) },
              message_ephemeral: true,
            },
          },
        ],
        edges: [
          ...options.map((option) => ({
            id: getEdgeId(),
            source: option.id,
            target: entryNodeId,
          })),
          {
            id: getEdgeId(),
            source: entryNodeId,
            target: errorHandlerNodeId,
          },
          ...actions.map((action, i) => ({
            id: getEdgeId(),
            source: i === 0 ? errorHandlerNodeId : actions[i - 1].id,
            sourceHandle: i === 0 ? "default" : undefined,
            target: action.id,
          })),
          {
            id: getEdgeId(),
            source: errorHandlerNodeId,
            sourceHandle: "error",
            target: errorResponseNodeId,
          },
        ],
      }),
    };
  }

  const statusServerNodeId = getNodeId("action_erlc_server_get");
  const playersServerNodeId = getNodeId("action_erlc_server_get");
  const announceRunNodeId = getNodeId("action_erlc_command_run");

  return {
    id: "erlc",
    version: 1,
    name: "ER:LC",
    description:
      "Show your ER:LC private server's status and players, and announce messages in it.",
    icon: SirenIcon,
    category: "games",
    setup: [
      "Enable ER:LC under Integrations and add your server key.",
      "Authorize Kite for your server once, so /announce can run commands.",
    ],
    inputs: [],
    commands: [
      command(
        "status",
        "Show the status of the ER:LC server.",
        [],
        [
          {
            id: statusServerNodeId,
            type: "action_erlc_server_get",
            data: {},
          },
          {
            id: getNodeId("action_response_create"),
            type: "action_response_create",
            data: {
              message_data: {
                content: `**{{ result('${statusServerNodeId}').Name }}**\nPlayers: {{ result('${statusServerNodeId}').CurrentPlayers }}/{{ result('${statusServerNodeId}').MaxPlayers }}\nJoin code: \`{{ result('${statusServerNodeId}').JoinKey }}\``,
              },
            },
          },
        ],
        (errorNodeId) =>
          `Couldn't get the server: {{ result('${errorNodeId}') }}`
      ),
      command(
        "players",
        "List the players in the ER:LC server.",
        [],
        [
          {
            id: playersServerNodeId,
            type: "action_erlc_server_get",
            data: { include_players: "true" },
          },
          {
            id: getNodeId("action_response_create"),
            type: "action_response_create",
            data: {
              message_data: {
                content: `**Players in {{ result('${playersServerNodeId}').Name }} ({{ result('${playersServerNodeId}').CurrentPlayers }}/{{ result('${playersServerNodeId}').MaxPlayers }})**\n{{ len(result('${playersServerNodeId}').Players) == 0 ? 'Nobody is in the server.' : join(map(result('${playersServerNodeId}').Players, split(#.Player, ':')[0] + ' (' + #.Team + ')'), '\\n') }}`,
              },
            },
          },
        ],
        (errorNodeId) =>
          `Couldn't get the players: {{ result('${errorNodeId}') }}`
      ),
      command(
        "announce",
        "Show a message to everyone in the ER:LC server.",
        [
          {
            id: getNodeId("option_command_argument"),
            type: "option_command_argument",
            data: {
              name: "message",
              description: "The message to show.",
              command_argument_type: "string",
              command_argument_required: true,
            },
          },
          {
            id: getNodeId("option_command_permissions"),
            type: "option_command_permissions",
            data: {
              // Manage Server
              command_permissions: "32",
            },
          },
        ],
        [
          {
            id: announceRunNodeId,
            type: "action_erlc_command_run",
            data: {
              erlc_command: ":m {{interaction.command.args.message}}",
            },
          },
          {
            id: getNodeId("action_response_create"),
            type: "action_response_create",
            data: {
              message_data: {
                content:
                  "Announced in the server:\n>>> {{interaction.command.args.message}}",
              },
              message_ephemeral: true,
            },
          },
        ],
        (errorNodeId) =>
          `Couldn't announce the message: {{ result('${errorNodeId}') }}`
      ),
    ],
    eventListeners: [],
  };
}
