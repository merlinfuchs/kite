import { describe, expect, it } from "vitest";
import { getDiscordApiOperation } from "../flow/discordApi";
import { integrationBlocks } from ".";

// What the service needs to run the blocks. Result schemas are only for the
// editor and the flow AI.
const serviceBlocks = integrationBlocks.map(({ result, ...block }) => ({
  type: block.type,
  integration: block.integration,
  operation: block.operation,
  method: block.method,
  path: block.path,
  credits: block.credits,
  audit_log_reason: !!block.audit_log_reason,
  fields: block.fields.map((f) => ({
    name: f.name,
    in: f.in,
    target: f.target ?? f.name,
    type: f.type,
    required: !!f.required,
    fallback: f.fallback ?? "",
    min: f.min ?? null,
    max: f.max ?? null,
    max_length: f.max_length ?? null,
  })),
  result: result ? { thing: result.thing ?? "", list: !!result.list } : null,
}));

describe("integration blocks", () => {
  it("match the file embedded in the service", async () => {
    await expect(
      JSON.stringify({ blocks: serviceBlocks }, null, 2) + "\n"
    ).toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/integration_blocks.json"
    );
  });

  it("have unique types and field names", () => {
    const types = integrationBlocks.map((b) => b.type);
    expect(new Set(types).size).toBe(types.length);
    for (const block of integrationBlocks) {
      const names = block.fields.map((f) => f.name);
      expect(new Set(names).size, block.type).toBe(names.length);
    }
  });

  it("match Discord's spec", () => {
    for (const block of integrationBlocks.filter(
      (b) => b.integration === "discord"
    )) {
      const op = getDiscordApiOperation(block.operation);
      expect(op, block.type).toBeDefined();
      expect([block.method, block.path], block.type).toEqual([
        op!.method,
        op!.path,
      ]);

      const pathFields = block.fields.filter((f) => f.in === "path");
      expect(
        pathFields.map((f) => f.target ?? f.name).sort(),
        block.type
      ).toEqual(op!.path_params.map((p) => p.name).sort());

      for (const field of block.fields.filter((f) => f.in === "query")) {
        const param = op!.query_params.find(
          (p) => p.name === (field.target ?? field.name)
        );
        expect(param, `${block.type}.${field.name}`).toBeDefined();
      }
      if (block.fields.some((f) => f.in === "body")) {
        expect(op!.has_body, block.type).toBe(true);
      }
    }
  });
});
