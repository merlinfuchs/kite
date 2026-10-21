import { Edge, Node } from "@xyflow/react";

// Example flows for the landing pages, rendered by FlowExample.
type Flow = { nodes: Node[]; edges: Edge[] };

function chain(nodes: Node[]): Edge[] {
  return nodes.slice(1).map((node, i) => ({
    id: `e${i}`,
    source: nodes[i].id,
    target: node.id,
    type: "fixed",
  }));
}

function column(
  steps: { type: string; data: Record<string, unknown> }[]
): Flow {
  const nodes = steps.map((step, i) => ({
    id: String(i),
    position: { x: 0, y: i * 150 },
    ...step,
  }));
  return { nodes, edges: chain(nodes) };
}

export const helloFlow = column([
  {
    type: "entry_command",
    data: { name: "hello", description: "Say hello to the bot" },
  },
  {
    type: "action_response_create",
    data: { message_data: { content: "Hello {{user.mention}}!" } },
  },
]);

export const banFlow: Flow = {
  nodes: [
    {
      id: "user",
      position: { x: -50, y: -150 },
      type: "option_command_argument",
      data: {
        name: "user",
        description: "The user you want to ban",
        command_argument_type: "user",
        command_argument_required: true,
      },
    },
    {
      id: "reason",
      position: { x: 250, y: -150 },
      type: "option_command_argument",
      data: {
        name: "reason",
        description: "Why you want to ban the user",
        command_argument_type: "string",
      },
    },
    {
      id: "ban",
      position: { x: 0, y: 0 },
      type: "entry_command",
      data: { name: "ban", description: "Ban a user" },
    },
    {
      id: "action",
      position: { x: 100, y: 150 },
      type: "action_member_ban",
      data: {
        user_target: "{{interaction.command.args.user.id}}",
        member_ban_delete_message_duration_seconds: "0",
      },
    },
    {
      id: "reply",
      position: { x: 0, y: 300 },
      type: "action_response_create",
      data: { message_data: { content: "Banned the user." } },
    },
  ],
  edges: [
    { id: "e1", source: "user", target: "ban", type: "fixed" },
    { id: "e2", source: "reason", target: "ban", type: "fixed" },
    { id: "e3", source: "ban", target: "action", type: "fixed" },
    { id: "e4", source: "action", target: "reply", type: "fixed" },
  ],
};

export const welcomeFlow = column([
  {
    type: "entry_event",
    data: {
      event_type: "guild_member_add",
      description: "Welcome new members",
    },
  },
  {
    type: "action_member_role_add",
    data: {
      user_target: "{{event.user.id}}",
      role_target: "123456789012345678",
    },
  },
  {
    type: "action_message_create",
    data: {
      channel_target: "123456789012345678",
      message_data: {
        content: "Welcome to the server, {{event.user.mention}}!",
      },
    },
  },
]);

export const askAiFlow = column([
  {
    type: "entry_command",
    data: { name: "ask", description: "Ask the AI a question" },
  },
  {
    type: "action_ai_chat_completion",
    data: {
      ai_chat_completion_data: {
        prompt: "{{interaction.command.args.question}}",
      },
    },
  },
  {
    type: "action_response_create",
    data: { message_data: { content: "{{nodes.1.result}}" } },
  },
]);

export const erlcStatusFlow = column([
  {
    type: "entry_command",
    data: { name: "status", description: "Show the ER:LC server's status" },
  },
  {
    type: "action_erlc_server_get",
    data: {},
  },
  {
    type: "action_response_create",
    data: {
      message_data: {
        content:
          "{{ result('1').CurrentPlayers }} players are in the server. Join with {{ result('1').JoinKey }}!",
      },
    },
  },
]);
