import { afterEach, describe, expect, it, vi } from "vitest";
import { KiteAPI, kiteAPI, tokenScopes } from "./api";

describe("kiteAPI", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("sends the token and the client IP", async () => {
    const fetch = vi.fn(async () => Response.json({ success: true, data: [] }));
    vi.stubGlobal("fetch", fetch);

    await kiteAPI("kite_x", "203.0.113.7")("/v1/apps");

    const [url, init] = fetch.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toMatch(/\/v1\/apps$/);
    expect(init.headers).toMatchObject({
      Authorization: "Bearer kite_x",
      "CF-Connecting-IP": "203.0.113.7",
    });
  });
});

describe("tokenScopes", () => {
  it("caches the scopes of a token unless asked for fresh ones", async () => {
    const api = vi.fn(async () => ({
      success: true as const,
      data: { scopes: ["kite:read"] },
    })) as unknown as KiteAPI;

    for (const fresh of [false, false, true]) {
      const res = await tokenScopes(api, "kite_cached", fresh);
      expect(res).toEqual({ success: true, data: ["kite:read"] });
    }
    expect(api).toHaveBeenCalledTimes(2);
  });
});

describe("tokenScopes of OAuth clients", () => {
  it("doesn't cache the scopes past the token's expiry", async () => {
    const api = vi.fn(async () => ({
      success: true as const,
      data: {
        scopes: ["kite:read"],
        access_expires_at: new Date(Date.now() - 1000).toISOString(),
      },
    })) as unknown as KiteAPI;

    await tokenScopes(api, "kite_expiring", false);
    await tokenScopes(api, "kite_expiring", false);
    expect(api).toHaveBeenCalledTimes(2);
  });
});
