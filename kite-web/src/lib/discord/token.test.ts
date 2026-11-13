import { describe, expect, it } from "vitest";
import { checkTokenFormat } from "./token";

describe("checkTokenFormat", () => {
  it("accepts bot tokens", () => {
    // Built from parts so it doesn't look like a leaked token.
    const token = ["MTA5ODc2NTQzMjEwOTg3NjU0Mw", "GaBcDe", "x".repeat(38)].join(
      "."
    );
    expect(checkTokenFormat(token)).toBeNull();
  });

  it("rejects the application ID", () => {
    expect(checkTokenFormat("1098765432109876543")).toBe("application_id");
  });

  it("rejects the public key", () => {
    expect(checkTokenFormat("a".repeat(64))).toBe("public_key");
  });

  it("rejects the client secret", () => {
    expect(checkTokenFormat("aB3_dE6-gH9iJ2kL5mN8oP1qR4sT7uV0")).toBe(
      "client_secret"
    );
  });
});
