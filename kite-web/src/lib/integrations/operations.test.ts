import { describe, expect, it } from "vitest";
import { allOperations } from "./operations";

describe("operations", () => {
  it("matches the file embedded in the service", async () => {
    await expect(
      JSON.stringify({ operations: allOperations() }, null, 2) + "\n"
    ).toMatchFileSnapshot("../../../../kite-service/pkg/flow/operations.json");
  });

  it("leaves out what would break how Kite manages an app", () => {
    // Registering commands would replace the ones Kite syncs, and editing
    // the application could change its interactions endpoint.
    const writes = allOperations().filter(
      (op) =>
        op.integration === "discord" &&
        op.method !== "GET" &&
        (op.path.startsWith("/applications/") || op.path.startsWith("/oauth2/"))
    );
    expect(writes.map((op) => op.id)).toEqual([]);
  });

  it("makes Discord reads free", () => {
    for (const op of allOperations()) {
      if (op.integration === "discord" && op.method === "GET") {
        expect(op.policy.credits, op.id).toBe(0);
      }
    }
  });
});
