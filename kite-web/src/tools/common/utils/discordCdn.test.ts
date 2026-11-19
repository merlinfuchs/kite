import { describe, expect, it } from "vitest";
import userAvatar from "./discordCdn";

describe("userAvatar", () => {
  it("uses GIFs for animated avatars", () => {
    expect(userAvatar({ id: "123", avatar: "a_animated" })).toBe(
      "https://cdn.discordapp.com/avatars/123/a_animated.gif"
    );
  });

  it("uses PNGs for static avatars", () => {
    expect(userAvatar({ id: "123", avatar: "static" })).toBe(
      "https://cdn.discordapp.com/avatars/123/static.png"
    );
  });
});
