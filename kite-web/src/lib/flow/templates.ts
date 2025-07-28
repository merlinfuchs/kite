import { Edge, Node } from "@xyflow/react";
import { getLayoutedElements } from "./layout";
import { getEdgeId, getNodeId } from "./nodes";
import { NodeData } from "./dataSchema";
import {
  BrainCircuitIcon,
  GavelIcon,
  LucideIcon,
  UserRoundPlusIcon,
} from "lucide-react";

export type Template = {
  name: string;
  description: string;
  icon: LucideIcon;
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
  return [getModerationTemplate(), getAITemplate(), getWelcomerTemplate()];
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
    flow.edges,
    {
      direction: "TB",
    }
  );
}

export function getModerationTemplate(): Template {
  const moderationBanEntryNodeId = getNodeId();
  const moderationBanOptionUserIdNodeId = getNodeId();
  const moderationBanOptionPermissionsNodeId = getNodeId();
  const moderationBanOptionReasonNodeId = getNodeId();
  const moderationBanActionMemberBanNodeId = getNodeId();
  const moderationBanActionResponseNodeId = getNodeId();

  const moderationUnbanEntryNodeId = getNodeId();
  const moderationUnbanOptionUserIdNodeId = getNodeId();
  const moderationUnbanOptionPermissionsNodeId = getNodeId();
  const moderationUnbanOptionReasonNodeId = getNodeId();
  const moderationUnbanActionMemberUnbanNodeId = getNodeId();
  const moderationUnbanActionResponseNodeId = getNodeId();

  const moderationKickEntryNodeId = getNodeId();
  const moderationKickOptionUserIdNodeId = getNodeId();
  const moderationKickOptionPermissionsNodeId = getNodeId();
  const moderationKickOptionReasonNodeId = getNodeId();
  const moderationKickActionMemberKickNodeId = getNodeId();
  const moderationKickActionResponseNodeId = getNodeId();

  const moderationMuteEntryNodeId = getNodeId();
  const moderationMuteOptionUserIdNodeId = getNodeId();
  const moderationMuteOptionPermissionsNodeId = getNodeId();
  const moderationMuteOptionDurationNodeId = getNodeId();
  const moderationMuteOptionReasonNodeId = getNodeId();
  const moderationMuteActionMemberTimeoutNodeId = getNodeId();
  const moderationMuteActionResponseNodeId = getNodeId();
  return {
    name: "Moderation",
    description:
      "A number of moderation commands to help you manage your server.",
    icon: GavelIcon,
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
  const aiAskCommandEntryNodeId = getNodeId();
  const aiAskCommandOptionQuestionNodeId = getNodeId();
  const aiAskCommandActionAiChatCompletionNodeId = getNodeId();
  const aiAskCommandActionResponseCreateNodeId = getNodeId();

  const aiAskEventEntryNodeId = getNodeId();
  const aiAskEventConditionNodeID = getNodeId();
  const aiAskEventConditionItemNodeId = getNodeId();
  const aiAskEventConditionItemElseNodeId = getNodeId();
  const aiAskEventActionAiChatCompletionNodeId = getNodeId();
  const aiAskEventActionMessageCreateNodeId = getNodeId();

  return {
    name: "Ask AI",
    description: "A command and event listener to let users ask AI questions.",
    icon: BrainCircuitIcon,
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
            {
              id: aiAskEventConditionNodeID,
              type: "control_condition_compare",
              data: {
                condition_base_value: "{{event.message.content}}",
              },
            },
            {
              id: aiAskEventConditionItemNodeId,
              type: "control_condition_item_compare",
              data: {
                condition_item_value: "{{app.user.mention}}",
                condition_item_mode: "contains",
              },
            },
            {
              id: aiAskEventConditionItemElseNodeId,
              type: "control_condition_item_else",
              data: {},
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
              id: aiAskEventActionMessageCreateNodeId,
              type: "action_message_create",
              data: {
                channel_target: "{{event.channel.id}}",
                message_data: {
                  content: `{{event.user.mention}} {{nodes.${aiAskEventActionAiChatCompletionNodeId}.result}}`,
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: aiAskEventEntryNodeId,
              target: aiAskEventConditionNodeID,
            },
            {
              id: getEdgeId(),
              source: aiAskEventConditionNodeID,
              target: aiAskEventConditionItemNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventConditionNodeID,
              target: aiAskEventConditionItemElseNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventConditionItemNodeId,
              target: aiAskEventActionAiChatCompletionNodeId,
            },
            {
              id: getEdgeId(),
              source: aiAskEventActionAiChatCompletionNodeId,
              target: aiAskEventActionMessageCreateNodeId,
            },
          ],
        }),
      },
    ],
  };
}

export function getWelcomerTemplate(): Template {
  const welcomerEntryNodeId = getNodeId();
  const welcomerActionMessageCreateNodeId = getNodeId();

  return {
    name: "Welcomer",
    description: "An event listener to welcome new users to the server.",
    icon: UserRoundPlusIcon,
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
