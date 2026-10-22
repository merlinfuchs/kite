import { describe, expect, it } from "vitest";
import { blockDataSchema } from ".";
import { discordEntryCommand } from "./discordEntryCommand";

const schema = blockDataSchema(discordEntryCommand);

function issues(data: Record<string, unknown>) {
  const res = schema.safeParse(data);
  return res.success
    ? {}
    : Object.fromEntries(
        res.error.issues.map((issue) => [issue.path.join("."), issue.message])
      );
}

describe("command entry", () => {
  it("checks the name and description of slash commands", () => {
    expect(issues({ name: "ticket open", description: "Open" })).toEqual({});
    expect(
      issues({ command_type: "chat_input", name: "ping", description: "Pong" })
    ).toEqual({});

    expect(
      issues({ name: "Report User", description: "Report" })
    ).toHaveProperty("name");
    expect(issues({ name: "a b c d", description: "Too deep" })).toHaveProperty(
      "name"
    );
    expect(issues({ name: "ping" })).toHaveProperty("description");
    expect(issues({ name: "ping", description: "" })).toHaveProperty(
      "description"
    );
    expect(issues({ name: "ping", description: "   " })).toHaveProperty(
      "description"
    );
  });

  it("only needs a name for user and message commands", () => {
    expect(issues({ command_type: "user", name: "Report User" })).toEqual({});
    expect(issues({ command_type: "message", name: "Quote It" })).toEqual({});
    expect(issues({ command_type: "user" })).toHaveProperty("name");
  });

  it("rejects unknown command types", () => {
    expect(issues({ command_type: "chat", name: "ping" })).toHaveProperty(
      "command_type"
    );
  });
});
