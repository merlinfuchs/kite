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
    expect(withoutIds(`${origin}/apps/123/constructor/abc`, origin)).toBe(
      `${origin}/apps/[appId]/constructor/abc`
    );
  });

  it("knows every page with an ID in apps", () => {
    const appDir = join(__dirname, "../pages/apps/[appId]");
    const dynamicDirs = readdirSync(appDir, {
      recursive: true,
      encoding: "utf8",
    })
      .filter((path) => path.endsWith("]"))
      .sort();
    const params = Object.entries(appPageParams).map(
      ([section, param]) => `${section}/${param}`
    );
    expect(dynamicDirs).toEqual(params.sort());
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
