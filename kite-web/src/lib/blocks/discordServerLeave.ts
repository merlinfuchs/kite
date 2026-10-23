import { BlockDefinition } from "./types";

// Always leaves the server the flow runs in. There is deliberately no server
// field, so the block can't be aimed at another server.
export const discordServerLeave: BlockDefinition = {
  type: "action_server_leave",
  id_prefix: "server",
  title: "Leave server",
  description: "Bot leaves the server the flow runs in",
  icon: "log-out",
  category: "Bot",
  requires: ["discord"],
  // Scheduled and webhook flows never run in a server.
  contexts: [
    "command",
    "component_button",
    "component_select_menu",
    "event_discord",
  ],
  credits: 1,
  strict_settings: true,
  run: { kind: "custom" },
};
