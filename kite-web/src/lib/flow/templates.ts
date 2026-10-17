import { Edge, Node } from "@xyflow/react";
import { getLayoutedElements } from "./layout";
import { getEdgeId, getNodeId } from "./nodes";
import { NodeData } from "./dataSchema";
import {
  BrainCircuitIcon,
  GavelIcon,
  InfoIcon,
  LaughIcon,
  LightbulbIcon,
  LucideIcon,
  MailIcon,
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
  return [
    getModerationTemplate(),
    getAITemplate(),
    getWelcomerTemplate(),
    getFunTemplate(),
    getSuggestionsTemplate(),
    getInfoTemplate(),
    getModmailTemplate(),
  ];
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

  const leaveEntryNodeId = getNodeId();
  const leaveActionMessageCreateNodeId = getNodeId();

  return {
    name: "Welcomer",
    description:
      "Event listeners to welcome new users to the server and say goodbye when they leave.",
    icon: UserRoundPlusIcon,
    inputs: [
      {
        key: "channel_id",
        label: "Channel ID",
        description: "The channel to send the welcome messages to.",
        type: "text",
        required: true,
      },
      {
        key: "leave_channel_id",
        label: "Leave Channel ID",
        description:
          "The channel to send the leave messages to. Uses the welcome channel if empty.",
        type: "text",
        required: false,
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
      {
        source: "discord",
        type: "guild_member_remove",
        description: "Say goodbye when a user leaves the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: leaveEntryNodeId,
              type: "entry_event",
              data: {
                event_type: "guild_member_remove",
                description: "Say goodbye when a user leaves the server.",
              },
            },
            {
              id: leaveActionMessageCreateNodeId,
              type: "action_message_create",
              data: {
                channel_target: inputs.leave_channel_id || inputs.channel_id,
                message_data: {
                  // The user already left, so a mention would show up as an
                  // unknown user for most members.
                  content:
                    "**{{event.user.display_name}}** has left the server. We now have {{guild.member_count}} members.",
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: leaveEntryNodeId,
              target: leaveActionMessageCreateNodeId,
            },
          ],
        }),
      },
    ],
  };
}

// apiCommand builds a command that sends a request to an external API and
// responds with an embed built from the response, or with an error if the API
// didn't answer with 200 OK.
function apiCommand({
  name,
  description,
  url,
  headers,
  embed,
}: {
  name: string;
  description: string;
  url: string;
  headers?: { key: string; value: string }[];
  // Gets the ID of the request block to read the response with
  // {{nodes.<id>.result.data}}.
  embed: (requestNodeId: string) => Record<string, unknown>;
}): Template["commands"][number] {
  const entryNodeId = getNodeId();
  const requestNodeId = getNodeId();
  const conditionNodeId = getNodeId();
  const conditionOkNodeId = getNodeId();
  const conditionElseNodeId = getNodeId();
  const responseNodeId = getNodeId();
  const errorResponseNodeId = getNodeId();

  return {
    name,
    description,
    flowSource: () => ({
      nodes: [
        {
          id: entryNodeId,
          type: "entry_command",
          data: {
            name,
            description,
          },
        },
        {
          id: requestNodeId,
          type: "action_http_request",
          data: {
            http_request_data: {
              url,
              method: "GET",
              headers,
            },
          },
        },
        {
          id: conditionNodeId,
          type: "control_condition_compare",
          data: {
            // Equal compares types too, so the number has to become text to
            // match the "200" of the condition item.
            condition_base_value: `{{string(nodes.${requestNodeId}.result.status_code)}}`,
          },
        },
        {
          id: conditionOkNodeId,
          type: "control_condition_item_compare",
          data: {
            condition_item_value: "200",
            condition_item_mode: "equal",
          },
        },
        {
          id: conditionElseNodeId,
          type: "control_condition_item_else",
          data: {},
        },
        {
          id: responseNodeId,
          type: "action_response_create",
          data: {
            message_data: {
              embeds: [embed(requestNodeId)],
            },
          },
        },
        {
          id: errorResponseNodeId,
          type: "action_response_create",
          data: {
            message_data: {
              content: "The API didn't answer, please try again later.",
            },
            message_ephemeral: true,
          },
        },
      ],
      edges: [
        {
          id: getEdgeId(),
          source: entryNodeId,
          target: requestNodeId,
        },
        {
          id: getEdgeId(),
          source: requestNodeId,
          target: conditionNodeId,
        },
        {
          id: getEdgeId(),
          source: conditionNodeId,
          target: conditionOkNodeId,
        },
        {
          id: getEdgeId(),
          source: conditionNodeId,
          target: conditionElseNodeId,
        },
        {
          id: getEdgeId(),
          source: conditionOkNodeId,
          target: responseNodeId,
        },
        {
          id: getEdgeId(),
          source: conditionElseNodeId,
          target: errorResponseNodeId,
        },
      ],
    }),
  };
}

export function getFunTemplate(): Template {
  return {
    name: "Fun",
    description:
      "Commands that get quotes, jokes and animal pictures from free APIs.",
    icon: LaughIcon,
    inputs: [],
    commands: [
      apiCommand({
        name: "quote",
        description: "Get a random inspiring quote.",
        url: "https://dummyjson.com/quotes/random/",
        embed: (id) => ({
          description: `> {{nodes.${id}.result.data.quote}}`,
          footer: {
            text: `— {{nodes.${id}.result.data.author}}`,
          },
          color: 0x5865f2,
        }),
      }),
      apiCommand({
        name: "joke",
        description: "Get a random dad joke.",
        url: "https://icanhazdadjoke.com/",
        // Without it the API responds with a web page instead of JSON.
        headers: [{ key: "Accept", value: "application/json" }],
        embed: (id) => ({
          title: "😂 Dad Joke",
          description: `{{nodes.${id}.result.data.joke}}`,
          color: 0xfee75c,
        }),
      }),
      apiCommand({
        name: "cat",
        description: "Get a random cat picture.",
        url: "https://api.thecatapi.com/v1/images/search",
        embed: (id) => ({
          title: "🐱 Meow!",
          image: {
            url: `{{nodes.${id}.result.data[0].url}}`,
          },
          color: 0xeb459e,
        }),
      }),
      apiCommand({
        name: "dog",
        description: "Get a random dog picture.",
        url: "https://dog.ceo/api/breeds/image/random",
        embed: (id) => ({
          title: "🐶 Woof!",
          image: {
            url: `{{nodes.${id}.result.data.message}}`,
          },
          color: 0x57f287,
        }),
      }),
    ],
    eventListeners: [],
  };
}

export function getSuggestionsTemplate(): Template {
  const entryNodeId = getNodeId();
  const optionSuggestionNodeId = getNodeId();
  const optionCooldownNodeId = getNodeId();
  const messageCreateNodeId = getNodeId();
  const reactionUpNodeId = getNodeId();
  const reactionDownNodeId = getNodeId();
  const responseNodeId = getNodeId();

  return {
    name: "Suggestions",
    description:
      "A command that posts suggestions to a channel for members to vote on.",
    icon: LightbulbIcon,
    inputs: [
      {
        key: "channel_id",
        label: "Channel ID",
        description: "The channel to post the suggestions in.",
        type: "text",
        required: true,
      },
    ],
    commands: [
      {
        name: "suggest",
        description: "Suggest something for the server.",
        flowSource: (inputs) => ({
          nodes: [
            {
              id: entryNodeId,
              type: "entry_command",
              data: {
                name: "suggest",
                description: "Suggest something for the server.",
              },
            },
            {
              id: optionSuggestionNodeId,
              type: "option_command_argument",
              data: {
                name: "suggestion",
                description: "What do you suggest?",
                command_argument_type: "string",
                command_argument_required: true,
              },
            },
            {
              id: optionCooldownNodeId,
              type: "option_command_cooldown",
              data: {
                cooldown_scope: "user",
                cooldown_duration_seconds: "60",
              },
            },
            {
              id: messageCreateNodeId,
              type: "action_message_create",
              data: {
                channel_target: inputs.channel_id,
                message_data: {
                  embeds: [
                    {
                      author: {
                        name: "{{user.display_name}}",
                        icon_url: "{{user.avatar_url}}",
                      },
                      title: "💡 New Suggestion",
                      description: "{{interaction.command.args.suggestion}}",
                      footer: {
                        text: "Vote with the reactions below",
                      },
                      color: 0xfee75c,
                    },
                  ],
                },
              },
            },
            {
              id: reactionUpNodeId,
              type: "action_message_reaction_create",
              data: {
                channel_target: inputs.channel_id,
                message_target: `{{nodes.${messageCreateNodeId}.result.id}}`,
                emoji_data: {
                  name: "👍",
                },
              },
            },
            {
              id: reactionDownNodeId,
              type: "action_message_reaction_create",
              data: {
                channel_target: inputs.channel_id,
                message_target: `{{nodes.${messageCreateNodeId}.result.id}}`,
                emoji_data: {
                  name: "👎",
                },
              },
            },
            {
              id: responseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  content: `Your suggestion has been posted in <#${inputs.channel_id}>!`,
                },
                message_ephemeral: true,
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: optionSuggestionNodeId,
              target: entryNodeId,
            },
            {
              id: getEdgeId(),
              source: optionCooldownNodeId,
              target: entryNodeId,
            },
            {
              id: getEdgeId(),
              source: entryNodeId,
              target: messageCreateNodeId,
            },
            {
              id: getEdgeId(),
              source: messageCreateNodeId,
              target: reactionUpNodeId,
            },
            {
              id: getEdgeId(),
              source: reactionUpNodeId,
              target: reactionDownNodeId,
            },
            {
              id: getEdgeId(),
              source: reactionDownNodeId,
              target: responseNodeId,
            },
          ],
        }),
      },
    ],
    eventListeners: [],
  };
}

export function getInfoTemplate(): Template {
  const avatarEntryNodeId = getNodeId();
  const avatarOptionUserNodeId = getNodeId();
  const avatarResponseNodeId = getNodeId();

  const userInfoEntryNodeId = getNodeId();
  const userInfoOptionUserNodeId = getNodeId();
  const userInfoResponseNodeId = getNodeId();

  const serverInfoEntryNodeId = getNodeId();
  const serverInfoContextsNodeId = getNodeId();
  const serverInfoResponseNodeId = getNodeId();

  // The picked user, or the user of the command if none was picked.
  const target = "(arg('user') ?? user)";

  return {
    name: "Server Info",
    description:
      "Commands that show the avatar and info of a user and info about the server.",
    icon: InfoIcon,
    inputs: [],
    commands: [
      {
        name: "avatar",
        description: "Show the avatar of a user.",
        flowSource: () => ({
          nodes: [
            {
              id: avatarEntryNodeId,
              type: "entry_command",
              data: {
                name: "avatar",
                description: "Show the avatar of a user.",
              },
            },
            {
              id: avatarOptionUserNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to show, yourself if empty.",
                command_argument_type: "user",
                command_argument_required: false,
              },
            },
            {
              id: avatarResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  embeds: [
                    {
                      title: `Avatar of {{${target}.display_name}}`,
                      url: `{{${target}.avatar_url}}`,
                      image: {
                        url: `{{${target}.avatar_url}}`,
                      },
                      color: 0x5865f2,
                    },
                  ],
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: avatarOptionUserNodeId,
              target: avatarEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: avatarEntryNodeId,
              target: avatarResponseNodeId,
            },
          ],
        }),
      },
      {
        name: "userinfo",
        description: "Show info about a user.",
        flowSource: () => ({
          nodes: [
            {
              id: userInfoEntryNodeId,
              type: "entry_command",
              data: {
                name: "userinfo",
                description: "Show info about a user.",
              },
            },
            {
              id: userInfoOptionUserNodeId,
              type: "option_command_argument",
              data: {
                name: "user",
                description: "The user to show, yourself if empty.",
                command_argument_type: "user",
                command_argument_required: false,
              },
            },
            {
              id: userInfoResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  embeds: [
                    {
                      author: {
                        name: `{{${target}.username}}`,
                        icon_url: `{{${target}.avatar_url}}`,
                      },
                      thumbnail: {
                        url: `{{${target}.avatar_url}}`,
                      },
                      description: `{{${target}.mention}}`,
                      fields: [
                        {
                          name: "ID",
                          value: `{{${target}.id}}`,
                          inline: true,
                        },
                        {
                          name: "Bot",
                          value: `{{${target}.is_bot ? 'Yes' : 'No'}}`,
                          inline: true,
                        },
                        {
                          name: "Account Created",
                          value: `<t:{{${target}.created_at}}:R>`,
                          inline: true,
                        },
                        {
                          name: "Joined Server",
                          value: `{{${target}.joined_at > 0 ? '<t:' + string(${target}.joined_at) + ':R>' : 'Not in this server'}}`,
                          inline: true,
                        },
                        {
                          name: "Avatar",
                          value: `{{${target}.avatar_url != '' ? '[Open full size](' + ${target}.avatar_url + ')' : 'Default avatar'}}`,
                          inline: true,
                        },
                        {
                          name: "Roles",
                          value: `{{${target}.role_count > 0 ? ${target}.role_mentions : 'None'}}`,
                        },
                      ],
                      color: 0x5865f2,
                    },
                  ],
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: userInfoOptionUserNodeId,
              target: userInfoEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: userInfoEntryNodeId,
              target: userInfoResponseNodeId,
            },
          ],
        }),
      },
      {
        name: "serverinfo",
        description: "Show info about the server.",
        flowSource: () => ({
          nodes: [
            {
              id: serverInfoEntryNodeId,
              type: "entry_command",
              data: {
                name: "serverinfo",
                description: "Show info about the server.",
              },
            },
            {
              id: serverInfoContextsNodeId,
              type: "option_command_contexts",
              data: {
                command_disabled_contexts: ["bot_dm", "private_channel"],
              },
            },
            {
              id: serverInfoResponseNodeId,
              type: "action_response_create",
              data: {
                message_data: {
                  embeds: [
                    {
                      title: "{{guild.name}}",
                      description: "{{guild.description}}",
                      thumbnail: {
                        url: "{{guild.icon_url}}",
                      },
                      fields: [
                        {
                          name: "Owner",
                          value: "<@{{guild.owner_id}}>",
                          inline: true,
                        },
                        {
                          name: "Members",
                          value: "{{guild.member_count}}",
                          inline: true,
                        },
                        {
                          name: "Created",
                          value: "<t:{{guild.created_at}}:R>",
                          inline: true,
                        },
                        {
                          name: "Channels",
                          value: "{{guild.channel_count}}",
                          inline: true,
                        },
                        {
                          name: "Roles",
                          value: "{{guild.role_count}}",
                          inline: true,
                        },
                        {
                          name: "Boosts",
                          value:
                            "{{guild.boost_count}} (Level {{guild.boost_level}})",
                          inline: true,
                        },
                      ],
                      footer: {
                        text: "ID: {{guild.id}}",
                      },
                      color: 0x5865f2,
                    },
                  ],
                },
              },
            },
          ],
          edges: [
            {
              id: getEdgeId(),
              source: serverInfoContextsNodeId,
              target: serverInfoEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: serverInfoEntryNodeId,
              target: serverInfoResponseNodeId,
            },
          ],
        }),
      },
    ],
    eventListeners: [],
  };
}

export function getModmailTemplate(): Template {
  const modmailEventEntryNodeId = getNodeId();
  const modmailEventForwardNodeId = getNodeId();
  const modmailEventConfirmNodeId = getNodeId();

  const modmailReplyEntryNodeId = getNodeId();
  const modmailReplyOptionUserNodeId = getNodeId();
  const modmailReplyOptionMessageNodeId = getNodeId();
  const modmailReplyOptionPermissionsNodeId = getNodeId();
  const modmailReplyActionSendNodeId = getNodeId();
  const modmailReplyActionResponseNodeId = getNodeId();
  const modmailReplyErrorHandlerNodeId = getNodeId();
  const modmailReplyErrorResponseNodeId = getNodeId();

  return {
    name: "Modmail",
    description:
      "Users DM the bot to reach your staff, and staff answer with /reply.",
    icon: MailIcon,
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
