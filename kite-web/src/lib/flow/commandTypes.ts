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

// How a command is named in lists, like /ping, or Report message for a User or
// Message command.
export function commandLabel(name: string, flow: FlowData) {
  return commandTypeOf(flow) === "chat_input" ? `/${name}` : name;
}

// The type of a command, from the entry block of its flow.
export function commandTypeOf(flow: FlowData): CommandType {
  const entry = flow.nodes?.find((n) => n.type === "entry_command");
  const type = entry?.data.command_type;
  return type === "user" || type === "message" ? type : "chat_input";
}

// The icon of a command's block in the editor, matching the icons of the
// command cards in the commands list. Names are those of lucide icons.
export function commandTypeIconName(data: { command_type?: string }) {
  switch (data.command_type) {
    case "user":
      return "user";
    case "message":
      return "message-square";
    default:
      return "square-slash";
  }
}
