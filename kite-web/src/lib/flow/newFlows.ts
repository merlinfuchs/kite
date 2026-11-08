import { getUniqueId } from "../utils";
import { getNodeId } from "./nodes";
import { FlowEdge, FlowNode } from "../types/flow.gen";

// The flows new commands, event listeners, buttons and select menus start
// with: only the entry block.

export type NewCommandType = "chat_input" | "user" | "message";
export type NewCommandInstall = "guild_install" | "user_install";

export function newCommandFlow(
  name: string,
  description: string,
  commandType: NewCommandType = "chat_input",
  install?: NewCommandInstall
) {
  const entryId = getNodeId("entry_command");
  const nodes: FlowNode[] = [
    {
      id: entryId,
      position: { x: 0, y: 0 },
      data: { name, description, command_type: commandType },
      type: "entry_command",
    },
  ];
  const edges: FlowEdge[] = [];

  if (install) {
    const contextsId = getNodeId("contexts");
    nodes.push({
      id: contextsId,
      position: { x: -250, y: 0 },
      data: {
        command_disabled_contexts:
          install === "guild_install"
            ? ["bot_dm", "private_channel"]
            : undefined,
        command_disabled_integrations: [
          install === "guild_install" ? "user_install" : "guild_install",
        ],
      },
      type: "option_command_contexts",
    });
    edges.push({
      id: `${contextsId}-${entryId}`,
      source: contextsId,
      target: entryId,
    });
  }

  return {
    nodes,
    edges,
  };
}

export function newEventListenerFlow(
  type: string,
  description: string,
  cron?: string
) {
  return {
    nodes: [
      {
        id: getNodeId("entry_event"),
        position: { x: 0, y: 0 },
        data: {
          event_type: type,
          event_schedule_cron: cron || undefined,
          description,
        },
        type: "entry_event",
      },
    ],
    edges: [],
  };
}

export function newComponentFlow() {
  return {
    nodes: [
      {
        id: getUniqueId().toString(),
        position: { x: 0, y: 0 },
        data: {},
        type: "entry_component_button",
      },
    ],
    edges: [],
  };
}
