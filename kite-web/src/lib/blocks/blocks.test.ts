import { describe, expect, it } from "vitest";
import { getDiscordApiOperation } from "../flow/discordApi";
import { getIntegration } from "../integrations";
import { blockDefinitions } from ".";

// What the service needs to run the blocks. Result schemas are only for the
// editor and the flow AI.
const serviceBlocks = blockDefinitions.map((block) => ({
  type: block.type,
  credits: block.credits,
  audit_log_reason: !!block.audit_log_reason,
  run: block.run,
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
  result: block.result
    ? { thing: block.result.thing ?? "", list: !!block.result.list }
    : null,
}));

describe("block definitions", () => {
  it("match the file embedded in the service", async () => {
    await expect(
      JSON.stringify({ blocks: serviceBlocks }, null, 2) + "\n"
    ).toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/block_definitions.json"
    );
  });

  it("have unique types and field names", () => {
    const types = blockDefinitions.map((b) => b.type);
    expect(new Set(types).size).toBe(types.length);
    for (const block of blockDefinitions) {
      const names = block.fields.map((f) => f.name);
      expect(new Set(names).size, block.type).toBe(names.length);
    }
  });

  it("use integrations that exist", () => {
    for (const block of blockDefinitions) {
      expect(getIntegration(block.run.integration), block.type).toBeDefined();
    }
  });

  it("match Discord's spec", () => {
    for (const block of blockDefinitions.filter(
      (b) => b.run.integration === "discord"
    )) {
      const op = getDiscordApiOperation(block.run.operation);
      expect(op, block.type).toBeDefined();
      expect([block.run.method, block.run.path], block.type).toEqual([
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
