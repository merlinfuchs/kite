import { describe, expect, it } from "vitest";
import { shortenResult } from "./discord";

const role = (i: number) => ({
  id: `${i}`,
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
    const shortened = shortenResult(result, "get_guild");
    expect(shortened).toMatchObject({
      truncated: true,
      result: { name: "Guild" },
    });
    const { roles } = (shortened as { result: { roles: unknown[] } }).result;
    expect(roles).toHaveLength(26);
    expect(roles[25]).toBe("... 175 more, get them with list_guild_roles");
  });

  it("shortens long strings", () => {
    const result = { id: "1", content: "x".repeat(50_000) };
    expect(shortenResult(result, "get_message")).toEqual({
      truncated: true,
      result: {
        id: "1",
        content: "x".repeat(1000) + "... 49000 more characters",
      },
    });
  });

  it("cuts the text when shortening isn't enough", () => {
    const result = Array.from({ length: 5 }, () =>
      Object.fromEntries(
        Array.from({ length: 500 }, (_, i) => [`field_${i}`, i])
      )
    );
    const shortened = shortenResult(result, "list_messages");
    expect(shortened).toMatchObject({ truncated: true });
    expect((shortened as { result: string }).result).toHaveLength(20_000);
  });
});
