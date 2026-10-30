import { getUniqueId } from "../utils";
import { getNodeId } from "./nodes";
import { CommandType, CommandTypeChatInput } from "../types/flow.gen";
import { isSlashCommand } from "./commandType";

// The flows new commands, event listeners, buttons and select menus start
// with: only the entry block.

export function newCommandFlow(
  name: string,
  description: string,
  commandType: CommandType = CommandTypeChatInput
) {
  return {
    nodes: [
      {
        id: getNodeId("entry_command"),
        position: { x: 0, y: 0 },
        data: {
          name,
          ...(isSlashCommand(commandType)
            ? { description }
            : { command_type: commandType }),
        },
        type: "entry_command",
      },
    ],
    edges: [],
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
