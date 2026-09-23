import { describe, expect, it } from "vitest";
import { toHTML } from "./discordMarkdown";

describe("toHTML", () => {
  it("escapes html in subtext", () => {
    const html = toHTML("-# <img src=x onerror=alert(1)> **bold**");

    expect(html).not.toContain("<img");
    expect(html).toContain("&lt;img");
    expect(html).toContain("<strong>bold</strong>");
  });
});
