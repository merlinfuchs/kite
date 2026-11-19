import { describe, expect, it } from "vitest";
import { discordAvatarUrl } from "./discordCdn";

describe("discordAvatarUrl", () => {
  it("uses GIFs for animated bot avatars", () => {
    expect(
      discordAvatarUrl(
        "https://cdn.discordapp.com/avatars/123/a_animated.png?size=256"
      )
    ).toBe("https://cdn.discordapp.com/avatars/123/a_animated.gif?size=256");
  });

  it("leaves non-Discord avatar URLs unchanged", () => {
    const url = "https://example.com/a_animated.png";
    expect(discordAvatarUrl(url)).toBe(url);
  });
});
