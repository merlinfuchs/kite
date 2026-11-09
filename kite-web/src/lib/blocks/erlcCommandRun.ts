import { BlockDefinition } from "./types";

export const erlcCommandRun: BlockDefinition = {
  type: "action_erlc_command_run",
  id_prefix: "erlc",
  title: "Run ER:LC command",
  description: "Run a command in your ER:LC private server",
  icon: "square-terminal",
  category: "ER:LC",
  strict_settings: true,
  run: {
    kind: "request",
    integration: "erlc",
    operation: "run_command",
  },
  fields: [
    {
      name: "erlc_command",
      in: "body",
      target: "command",
      type: "string",
      label: "Command",
      description:
        'The command to run, like ":h Hello everyone!". ER:LC runs one command every 5 seconds per server, so the block waits for its turn, up to 10 seconds.',
      required: true,
    },
  ],
};
