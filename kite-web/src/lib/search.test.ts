import { describe, expect, it } from "vitest";
import { matchesSearch } from "./search";

describe("matchesSearch", () => {
  it("matches everything for an empty query", () => {
    expect(matchesSearch("", ["ban"])).toBe(true);
    expect(matchesSearch("   ", [])).toBe(true);
  });

  it("is case-insensitive", () => {
    expect(matchesSearch("BAN", ["user-ban"])).toBe(true);
  });

  it("requires every term but in any field or order", () => {
    expect(matchesSearch("ban user", ["user-ban", "Bans a member"])).toBe(true);
    expect(matchesSearch("ban kick", ["user-ban", "Bans a member"])).toBe(
      false
    );
  });

  it("ignores empty fields", () => {
    expect(matchesSearch("x", [null, undefined, ""])).toBe(false);
  });
});
