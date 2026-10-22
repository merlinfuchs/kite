import { describe, expect, it } from "vitest";
import {
  buildKiteComponentEmbed,
  serializeComponentEmbed,
} from "./discordComponentEmbed";

const BUTTON_KEYS = [
  "id",
  "type",
  "url",
  "style",
  "label",
  "emoji",
  "disabled",
];

function walk(c: any, visit: (c: any) => void) {
  visit(c);
  for (const child of c.components ?? []) walk(child, visit);
  if (c.accessory) walk(c.accessory, visit);
}

describe("discord component embed", () => {
  const payload = buildKiteComponentEmbed({
    title: "Free Discord Bot Maker, No Coding Required",
    description:
      "Make your own Discord bot for free with Kite. Build slash commands, buttons and events in a visual editor, no coding required. Hosted 24/7.",
    url: "https://kite.onl/",
  });

  it("stays within Discord's size limit", () => {
    expect(serializeComponentEmbed(payload).length).toBeLessThanOrEqual(3000);
  });

  it("is a single container with at most 40 components", () => {
    expect(Object.keys(payload)).toEqual(["component"]);
    expect(payload.component.type).toBe(17);
    let count = 0;
    walk(payload.component, () => count++);
    expect(count).toBeLessThanOrEqual(40);
  });

  it("only uses link buttons with allowed keys", () => {
    walk(payload.component, (c) => {
      if (c.type !== 2) return;
      expect(c.style).toBe(5);
      expect(c.url).toMatch(/^https?:\/\//);
      for (const key of Object.keys(c)) expect(BUTTON_KEYS).toContain(key);
    });
  });

  it("uses absolute media URLs", () => {
    walk(payload.component, (c) => {
      const urls = [
        c.media?.url,
        ...(c.items ?? []).map((i: any) => i.media.url),
      ];
      for (const url of urls.filter(Boolean))
        expect(url).toMatch(/^https?:\/\//);
    });
  });
});
