import { afterEach, describe, expect, it, vi } from "vitest";
import { KiteAPI, kiteAPI, tokenAccess } from "./api";

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

describe("tokenAccess", () => {
  it("caches the scopes of a token unless asked for fresh ones", async () => {
    const api = vi.fn(async () => ({
      success: true as const,
      data: { scopes: ["kite:read"] },
    })) as unknown as KiteAPI;

    for (const fresh of [false, false, true]) {
      const res = await tokenAccess(api, "kite_cached", fresh);
      expect(res).toEqual({
        success: true,
        data: { scopes: ["kite:read"], oauth: false },
      });
    }
    expect(api).toHaveBeenCalledTimes(2);
  });
});

describe("tokenAccess of OAuth clients", () => {
  it("doesn't cache the scopes past the token's expiry", async () => {
    const api = vi.fn(async () => ({
      success: true as const,
      data: {
        scopes: ["kite:read"],
        access_expires_at: new Date(Date.now() - 1000).toISOString(),
      },
    })) as unknown as KiteAPI;

    await tokenAccess(api, "kite_expiring", false);
    await tokenAccess(api, "kite_expiring", false);
    expect(api).toHaveBeenCalledTimes(2);
  });
});

describe("kiteAPI with a deleted token", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("forgets its cached scopes", async () => {
    const cached = vi.fn(async () => ({
      success: true as const,
      data: { scopes: ["kite:read"], access_expires_at: null },
    })) as unknown as KiteAPI;
    await tokenAccess(cached, "kite_deleted", false);

    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json({
          success: false,
          error: { code: "invalid_api_token", message: "invalid" },
        })
      )
    );
    const api = kiteAPI("kite_deleted");
    await api("/v1/apps");

    const res = await tokenAccess(api, "kite_deleted", false);
    expect(res.success).toBe(false);
  });
});
