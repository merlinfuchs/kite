import { describe, expect, it } from "vitest";
import { newCommandFlow } from "./newFlows";

describe("newCommandFlow", () => {
  it("creates a slash command with a server-only install option", () => {
    const flow = newCommandFlow(
      "help",
      "Shows help",
      "chat_input",
      "guild_install"
    );
    const entry = flow.nodes.find((node) => node.type === "entry_command");
    const contexts = flow.nodes.find(
      (node) => node.type === "option_command_contexts"
    );

    expect(entry?.data.command_type).toBe("chat_input");
    expect(contexts?.data.command_disabled_contexts).toEqual([
      "bot_dm",
      "private_channel",
    ]);
    expect(contexts?.data.command_disabled_integrations).toEqual([
      "user_install",
    ]);
    expect(flow.edges).toEqual([
      {
        id: `${contexts?.id}-${entry?.id}`,
        source: contexts?.id,
        target: entry?.id,
      },
    ]);
  });

  it("creates a contextual command with a user install option", () => {
    const flow = newCommandFlow("inspect", "", "user", "user_install");
    const entry = flow.nodes.find((node) => node.type === "entry_command");
    const contexts = flow.nodes.find(
      (node) => node.type === "option_command_contexts"
    );

    expect(entry?.data.command_type).toBe("user");
    expect(contexts?.data.command_disabled_integrations).toEqual([
      "guild_install",
    ]);
    expect(flow.edges).toHaveLength(1);
  });

  it("creates a message context menu command", () => {
    const flow = newCommandFlow("summarize", "", "message");
    const entry = flow.nodes.find((node) => node.type === "entry_command");

    expect(entry?.data.command_type).toBe("message");
  });
});
