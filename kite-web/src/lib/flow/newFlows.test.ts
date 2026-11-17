import { describe, expect, it } from "vitest";
import { newCommandFlow } from "./newFlows";

describe("newCommandFlow", () => {
  it("creates a slash command", () => {
    const flow = newCommandFlow("help", "Shows help");

    expect(flow.nodes).toHaveLength(1);
    expect(flow.nodes[0].data).toEqual({
      name: "help",
      description: "Shows help",
    });
    expect(flow.edges).toEqual([]);
  });

  it("creates a User command", () => {
    const flow = newCommandFlow("Inspect", "", "user");

    expect(flow.nodes).toHaveLength(1);
    expect(flow.nodes[0].data).toEqual({
      name: "Inspect",
      command_type: "user",
    });
  });
});
