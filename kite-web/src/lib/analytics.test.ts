import { readdirSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { appPageParams, withoutIds } from "./analytics";

const origin = "https://kite.onl";

describe("withoutIds", () => {
  it("replaces IDs in app pages", () => {
    expect(withoutIds(`${origin}/apps/123`, origin)).toBe(
      `${origin}/apps/[appId]`
    );
    expect(withoutIds(`${origin}/apps/123/commands`, origin)).toBe(
      `${origin}/apps/[appId]/commands`
    );
    expect(withoutIds(`${origin}/apps/123/commands/456?tab=1#x`, origin)).toBe(
      `${origin}/apps/[appId]/commands/[cmdId]`
    );
    expect(withoutIds(`${origin}/apps/123/marketplace/abc`, origin)).toBe(
      `${origin}/apps/[appId]/marketplace/[templateId]`
    );
  });

  it("knows every page with an ID in apps", () => {
    const appDir = join(__dirname, "../pages/apps/[appId]");
    const params = Object.fromEntries(
      readdirSync(appDir, { withFileTypes: true })
        .filter((section) => section.isDirectory())
        .flatMap((section) =>
          readdirSync(join(appDir, section.name))
            .filter((name) => name.startsWith("["))
            .map((name) => [section.name, name])
        )
    );
    expect(appPageParams).toEqual(params);
  });

  it("keeps other pages", () => {
    for (const url of [
      `${origin}/apps`,
      `${origin}/marketplace/abc?utm_source=x`,
      "https://example.com/apps/123",
      "",
    ]) {
      expect(withoutIds(url, origin)).toBe(url);
    }
  });
});
