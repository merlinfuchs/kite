import { getUniqueId } from "../utils";
import { getNodeId } from "./nodes";
import { FlowEdge, FlowNode } from "../types/flow.gen";
import { CommandType } from "./commandTypes";

// The flows new commands, event listeners, buttons and select menus start
// with: only the entry block, and a Command Contexts block for commands that
// don't work everywhere.

// Where a new command can be used. Commands work everywhere unless their
// Command Contexts block limits them.
export type CommandAvailability = "everywhere" | "servers" | "users";

export function newCommandFlow(
  name: string,
  description: string,
  commandType: CommandType = "chat_input",
  availability: CommandAvailability = "everywhere"
) {
  const entryId = getNodeId("entry_command");
  const nodes: FlowNode[] = [
    {
      id: entryId,
      position: { x: 0, y: 0 },
      data:
        commandType === "chat_input"
          ? { name, description }
          : { name, command_type: commandType },
      type: "entry_command",
    },
  ];
  const edges: FlowEdge[] = [];

  if (availability !== "everywhere") {
    const contextsId = getNodeId("option_command_contexts");
    nodes.push({
      id: contextsId,
      position: { x: 0, y: -200 },
      data:
        availability === "servers"
          ? {
              command_disabled_contexts: ["bot_dm", "private_channel"],
              command_disabled_integrations: ["user_install"],
            }
          : { command_disabled_integrations: ["guild_install"] },
      type: "option_command_contexts",
    });
    edges.push({
      id: getUniqueId().toString(),
      source: contextsId,
      target: entryId,
    });
  }

  return { nodes, edges };
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
