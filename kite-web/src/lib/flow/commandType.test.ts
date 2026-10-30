import { describe, expect, it } from "vitest";
import { commandNameLabel, commandTypeIcon } from "./commandType";

describe("command labels and icons", () => {
  it("uses a slash prefix and icon for slash commands by default", () => {
    expect(commandNameLabel("hello")).toBe("/hello");
    expect(commandTypeIcon()).toBe("square-slash");
  });

  it("does not use slash styling for user context commands", () => {
    expect(commandNameLabel("View User", "user")).toBe("View User");
    expect(commandTypeIcon("user")).toBe("user-round");
  });

  it("does not use slash styling for message context commands", () => {
    expect(commandNameLabel("Summarize", "message")).toBe("Summarize");
    expect(commandTypeIcon("message")).toBe("message-square");
  });
});
