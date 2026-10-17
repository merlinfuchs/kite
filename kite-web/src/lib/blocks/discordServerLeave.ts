import { BlockDefinition } from "./types";

// Always leaves the server the flow runs in. There is deliberately no server
// field, so a flow can't make the app leave a server other than its own.
export const discordServerLeave: BlockDefinition = {
  type: "action_server_leave",
  title: "Leave Server",
  description: "Bot leaves the server the flow runs in",
  icon: "log-out",
  category: "Bot",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [],
  run: { kind: "custom" },
};
