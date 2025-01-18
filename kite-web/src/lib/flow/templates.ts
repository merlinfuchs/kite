import { Edge, Node } from "@xyflow/react";
import { getLayoutedElements } from "./layout";
import { getEdgeId, getNodeId } from "./nodes";
import { NodeData } from "./data";
import { BrainCircuitIcon, GavelIcon, LucideIcon } from "lucide-react";

export type Template = {
  name: string;
  description: string;
  icon: LucideIcon;
  commands: {
    name: string;
    description: string;
    flow_source: {
      nodes: Omit<Node<NodeData>, "position">[];
      edges: Edge[];
    };
  }[];
  eventListeners: {
    source: string;
    type: string;
    description: string;
    flow_source: {
      nodes: Omit<Node<NodeData>, "position">[];
      edges: Edge[];
    };
  }[];
};

export function getTemplates() {
  return [getModerationTemplate(), getAITemplate()];
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
  const moderationBanOptionReasonNodeId = getNodeId();
  const moderationBanActionMemberBanNodeId = getNodeId();

  return {
    name: "Moderation",
    description:
      "A number of moderation commands to help you manage your server.",
    icon: GavelIcon,
    commands: [
      {
        name: "ban",
        description: "Ban a user from the server.",
        flow_source: {
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
              id: moderationBanActionMemberBanNodeId,
              type: "action_member_ban",
              data: {
                member_target: "{{interaction.command.args.user}}",
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
              source: moderationBanOptionReasonNodeId,
              target: moderationBanEntryNodeId,
            },
            {
              id: getEdgeId(),
              source: moderationBanEntryNodeId,
              target: moderationBanActionMemberBanNodeId,
            },
          ],
        },
      },
      {
        name: "kick",
        description: "Kick a user from the server.",
        flow_source: {
          nodes: [],
          edges: [],
        },
      },
      {
        name: "mute",
        description: "Mute a user in the server.",
        flow_source: {
          nodes: [],
          edges: [],
        },
      },
    ],
    eventListeners: [],
  };
}

export function getAITemplate(): Template {
  return {
    name: "AI",
    description: "A number of AI commands to help you manage your server.",
    icon: BrainCircuitIcon,
    commands: [
      {
        name: "ask",
        description: "Ask a question to the AI.",
        flow_source: {
          nodes: [],
          edges: [],
        },
      },
    ],
    eventListeners: [
      {
        source: "discord",
        type: "message_create",
        description: "Ask a question to the AI by pinging the bot.",
        flow_source: {
          nodes: [],
          edges: [],
        },
      },
    ],
  };
}
