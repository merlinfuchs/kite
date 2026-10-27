import { describe, expect, it } from "vitest";
import { shortenResult } from "./discord";

const role = (i: number) => ({
  id: String(1000000000000000000 + i),
  name: `Role ${i}`,
  description: "A role of the server, like the ones Discord returns",
  permissions: "6445981249",
  colors: { primary_color: 0, secondary_color: null, tertiary_color: null },
  hoist: false,
  managed: false,
  mentionable: false,
});

describe("shortenResult", () => {
  it("keeps small results as they are", () => {
    const result = { id: "1", name: "Guild", roles: [role(1), role(2)] };
    expect(shortenResult(result, "get_guild")).toBe(result);
  });

  it("keeps long lists of small results", () => {
    const result = Array.from({ length: 100 }, (_, i) => ({ id: String(i) }));
    expect(shortenResult(result, "list_guild_members")).toBe(result);
  });

  it("shortens big lists and names their endpoint", () => {
    const result = {
      id: "1",
      name: "Guild",
      roles: Array.from({ length: 200 }, (_, i) => role(i)),
    };
    const shortened = shortenResult(result, "get_guild") as {
      truncated: boolean;
      result: { name: string; roles: unknown[] };
    };
    expect(shortened.truncated).toBe(true);
    expect(shortened.result.name).toBe("Guild");
    expect(shortened.result.roles).toHaveLength(26);
    expect(shortened.result.roles[25]).toBe(
      "... 175 more, get them with list_guild_roles"
    );
  });

  it("cuts the text when shortening lists isn't enough", () => {
    const result = { content: "x".repeat(50_000) };
    const shortened = shortenResult(result, "get_message") as {
      truncated: boolean;
      result: string;
    };
    expect(shortened.truncated).toBe(true);
    expect(shortened.result).toHaveLength(20_000);
  });
});
