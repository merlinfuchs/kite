import { afterEach, describe, expect, it, vi } from "vitest";
import { kiteAPI } from "./api";

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
