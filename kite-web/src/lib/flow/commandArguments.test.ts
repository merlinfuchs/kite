import { describe, expect, it } from "vitest";
import { getCommandArguments, moveCommandArgument } from "./commandArguments";
import { testEdge, testNode } from "./testUtils";

function flow(order?: string[]) {
  const entry = testNode("entry", "entry_command", {
    command_argument_order: order,
  });
  const nodes = [
    entry,
    testNode("reason", "option_command_argument"),
    testNode("user", "option_command_argument", {
      command_argument_required: true,
    }),
    testNode("days", "option_command_argument"),
    testNode("duration", "option_command_argument", {
      command_argument_required: true,
    }),
    testNode("permissions", "option_command_permissions"),
    testNode("unconnected", "option_command_argument"),
  ];
  const edges = [
    testEdge("reason", "entry"),
    testEdge("user", "entry"),
    testEdge("permissions", "entry"),
    testEdge("days", "entry"),
    testEdge("duration", "entry"),
  ];

  return getCommandArguments(entry, nodes, edges);
}

describe("getCommandArguments", () => {
  it("uses the connection order without an order", () => {
    expect(flow().map((n) => n.id)).toEqual([
      "duration",
      "user",
      "reason",
      "days",
    ]);
  });

  it("follows the order", () => {
    expect(
      flow(["user", "duration", "days", "reason"]).map((n) => n.id)
    ).toEqual(["user", "duration", "days", "reason"]);
  });

  it("keeps required arguments first", () => {
    expect(
      flow(["days", "user", "reason", "duration"]).map((n) => n.id)
    ).toEqual(["user", "duration", "days", "reason"]);
  });

  it("puts arguments missing from the order last", () => {
    expect(flow(["days", "user", "deleted"]).map((n) => n.id)).toEqual([
      "user",
      "duration",
      "days",
      "reason",
    ]);
  });
});

describe("moveCommandArgument", () => {
  it("swaps an argument with its neighbour", () => {
    expect(moveCommandArgument(flow(), 0, 1)).toEqual([
      "user",
      "duration",
      "reason",
      "days",
    ]);
    expect(moveCommandArgument(flow(), 3, -1)).toEqual([
      "duration",
      "user",
      "days",
      "reason",
    ]);
  });

  it("doesn't move past the first or last argument", () => {
    expect(moveCommandArgument(flow(), 0, -1)).toBeNull();
    expect(moveCommandArgument(flow(), 3, 1)).toBeNull();
  });

  it("doesn't move required arguments below optional ones", () => {
    expect(moveCommandArgument(flow(), 1, 1)).toBeNull();
    expect(moveCommandArgument(flow(), 2, -1)).toBeNull();
  });
});
