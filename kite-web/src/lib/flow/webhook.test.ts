import { describe, expect, it } from "vitest";
import { absoluteUrl } from "./webhook";

describe("absoluteUrl", () => {
  it("adds the origin to relative URLs", () => {
    expect(
      absoluteUrl("/v1/apps/a/webhooks/b/c", "http://localhost:8080")
    ).toBe("http://localhost:8080/v1/apps/a/webhooks/b/c");
  });

  it("keeps absolute URLs", () => {
    expect(
      absoluteUrl("https://api.kite.onl/v1/apps/a", "http://localhost:8080")
    ).toBe("https://api.kite.onl/v1/apps/a");
  });
});
