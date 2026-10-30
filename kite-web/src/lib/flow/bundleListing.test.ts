import { describe, expect, it } from "vitest";
import { bundleTagPattern, normalizeTag } from "./bundleListing";

describe("normalizeTag", () => {
  it("turns what was typed into a tag", () => {
    expect(normalizeTag("Roblox")).toBe("roblox");
    expect(normalizeTag("  Ticket System ")).toBe("ticket-system");
    expect(normalizeTag("role_drop")).toBe("role-drop");
    expect(normalizeTag("--a--b--")).toBe("a-b");
    expect(normalizeTag("🎮 games!")).toBe("games");
    expect(normalizeTag("🎮")).toBe("");
  });

  it("only makes tags the service accepts", () => {
    for (const text of ["Economy", "a very long tag name over twenty", "x-"]) {
      const tag = normalizeTag(text);
      expect(tag.length).toBeLessThanOrEqual(20);
      if (tag) expect(tag).toMatch(bundleTagPattern);
    }
  });
});
