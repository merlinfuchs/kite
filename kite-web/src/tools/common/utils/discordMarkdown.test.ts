import { describe, expect, it } from "vitest";
import { toHTML } from "./discordMarkdown";

describe("toHTML", () => {
  it("escapes html in subtext", () => {
    const html = toHTML("-# <img src=x onerror=alert(1)> **bold**");

    expect(html).not.toContain("<img");
    expect(html).toContain("&lt;img");
    expect(html).toContain("<strong>bold</strong>");
  });

  it("renders timestamps, also ones filled in when the message is sent", () => {
    expect(toHTML("<t:1700000000:d>")).toContain("discord-timestamp");
    expect(toHTML("<t:{{ now().Unix() }}:R>")).toContain("0 seconds");
    expect(toHTML("<t:{{now().Unix() + 7200}}:R>")).toContain("2 hours");
    expect(toHTML("<t:99999999999999999:f>")).toContain("Invalid Date");
  });
});
