import { describe, expect, it } from "vitest";
import { newCommandFlow } from "./newFlows";

describe("newCommandFlow", () => {
  it("creates a slash command that works everywhere", () => {
    const flow = newCommandFlow("help", "Shows help");

    expect(flow.nodes).toHaveLength(1);
    expect(flow.nodes[0].data).toEqual({
      name: "help",
      description: "Shows help",
    });
    expect(flow.edges).toEqual([]);
  });

  it("limits a command to servers", () => {
    const flow = newCommandFlow("help", "Shows help", "chat_input", "servers");
    const entry = flow.nodes.find((node) => node.type === "entry_command");
    const contexts = flow.nodes.find(
      (node) => node.type === "option_command_contexts"
    );

    expect(contexts?.data.command_disabled_contexts).toEqual([
      "bot_dm",
      "private_channel",
    ]);
    expect(contexts?.data.command_disabled_integrations).toEqual([
      "user_install",
    ]);
    expect(flow.edges).toMatchObject([
      { source: contexts?.id, target: entry?.id },
    ]);
  });

  it("creates a User command for users who added the bot", () => {
    const flow = newCommandFlow("Inspect", "", "user", "users");
    const entry = flow.nodes.find((node) => node.type === "entry_command");
    const contexts = flow.nodes.find(
      (node) => node.type === "option_command_contexts"
    );

    expect(entry?.data).toEqual({ name: "Inspect", command_type: "user" });
    expect(contexts?.data.command_disabled_integrations).toEqual([
      "guild_install",
    ]);
    expect(flow.edges).toHaveLength(1);
  });
});
