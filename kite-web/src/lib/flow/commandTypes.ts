import { FlowData } from "../types/flow.gen";

// Slash commands, and User and Message commands, which show in the Apps menu
// when someone right-clicks a user or message. Commands saved before there
// were types have none and are slash commands.
export type CommandType = "chat_input" | "user" | "message";

export const commandTypeOptions: { value: CommandType; label: string }[] = [
  { value: "chat_input", label: "Slash command" },
  { value: "user", label: "User command" },
  { value: "message", label: "Message command" },
];

export function isContextMenuCommand(data: { command_type?: string }) {
  return data.command_type === "user" || data.command_type === "message";
}

// The type of a command, from the entry block of its flow.
export function commandTypeOf(flow: FlowData): CommandType {
  const entry = flow.nodes?.find((n) => n.type === "entry_command");
  const type = entry?.data.command_type;
  return type === "user" || type === "message" ? type : "chat_input";
}
