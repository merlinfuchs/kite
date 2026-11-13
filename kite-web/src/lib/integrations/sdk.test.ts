import { describe, expect, it } from "vitest";
import { codeDeclarations, codeOperations, codeSignature } from "./sdk";

describe("sdk", () => {
  it("matches the kite.d.ts of the docs", async () => {
    await expect(codeDeclarations()).toMatchFileSnapshot(
      "../../../../kite-docs/static/kite.d.ts"
    );
  });

  it("writes calls as methods", () => {
    const op = codeOperations().find(
      (o) => o.namespace === "discord" && o.name === "listMessages"
    )!;
    expect(codeSignature(op)).toContain("channel_id: string;");
    expect(codeSignature(op)).toContain("limit?: number;");
    expect(codeSignature(op)).toContain("Costs nothing.");
  });
});
