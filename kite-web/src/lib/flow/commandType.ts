import {
  CommandType,
  CommandTypeChatInput,
  CommandTypeMessage,
  CommandTypeUser,
  FlowData,
} from "../types/flow.gen";

export const commandTypeOptions = [
  { value: CommandTypeChatInput, label: "Slash Command" },
  { value: CommandTypeUser, label: "User Context Menu" },
  { value: CommandTypeMessage, label: "Message Context Menu" },
];

// Commands without a type are slash commands, like in the service.
export function isSlashCommand(commandType?: CommandType) {
  return !commandType || commandType === CommandTypeChatInput;
}

export function commandType(flow: FlowData): CommandType {
  const entry = flow.nodes?.find((n) => n.type === "entry_command");
  return entry?.data.command_type || CommandTypeChatInput;
}
