import { describe, expect, it } from "vitest";
import { getIntegration, integrations } from "../integrations";
import { blockDefinitions, blockIntegrations, requestBlocks } from ".";

// What the service needs: how each block runs, which integrations it needs
// and the fields of request blocks. Schemas are only for the editor and the
// flow AI, and credits that depend on the settings are computed in Go.
const serviceBlocks = blockDefinitions.map((block) => ({
  type: block.type,
  credits: typeof block.credits === "number" ? block.credits : null,
  audit_log_reason: !!block.audit_log_reason,
  requires: blockIntegrations(block),
  run: block.run,
  fields: (block.fields ?? []).map((f) => ({
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

// The service doesn't read the APIs.
const serviceIntegrations = integrations.map(({ api: _, ...rest }) => rest);

describe("block definitions", () => {
  it("match the file embedded in the service", async () => {
    await expect(
      JSON.stringify(
        { integrations: serviceIntegrations, blocks: serviceBlocks },
        null,
        2
      ) + "\n"
    ).toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/block_definitions.json"
    );
  });

  it("have unique types and field names", () => {
    const types = blockDefinitions.map((b) => b.type);
    expect(new Set(types).size).toBe(types.length);
    for (const block of requestBlocks()) {
      const names = block.fields.map((f) => f.name);
      expect(new Set(names).size, block.type).toBe(names.length);
    }
  });

  it("send requests with fields and a fixed cost", () => {
    for (const block of blockDefinitions.filter(
      (b) => b.run.kind === "request"
    )) {
      expect(block.fields, block.type).toBeDefined();
      expect(typeof block.credits, block.type).toBe("number");
    }
  });

  it("describe the fields of blocks without a schema", () => {
    for (const block of requestBlocks().filter((b) => !b.schema)) {
      for (const field of block.fields) {
        expect(field.label, `${block.type}.${field.name}`).toBeTruthy();
        expect(field.description, `${block.type}.${field.name}`).toBeTruthy();
      }
    }
  });

  it("use integrations that exist", () => {
    for (const block of blockDefinitions) {
      for (const id of blockIntegrations(block)) {
        expect(getIntegration(id), `${block.type} ${id}`).toBeDefined();
      }
    }
  });

  it("match their integration's API", () => {
    for (const block of requestBlocks()) {
      const op = getIntegration(block.run.integration)?.api?.operations.find(
        (o) => o.id === block.run.operation
      );
      expect(op, block.type).toBeDefined();
      if (!op) continue;

      expect([block.run.method, block.run.path], block.type).toEqual([
        op.method,
        op.path,
      ]);

      const targets = (location: string) =>
        block.fields
          .filter((f) => f.in === location)
          .map((f) => f.target ?? f.name);
      expect(targets("path").sort(), block.type).toEqual(
        op.path_params.map((p) => p.name).sort()
      );
      for (const name of targets("query")) {
        expect(
          op.query_params.some((p) => p.name === name),
          `${block.type}.${name}`
        ).toBe(true);
      }

      if (targets("body").length > 0) {
        expect(op.has_body, block.type).toBe(true);
      }
      if (op.body_params) {
        for (const name of targets("body")) {
          expect(
            op.body_params.some((p) => p.name === name),
            `${block.type}.${name}`
          ).toBe(true);
        }
        for (const param of op.body_params.filter((p) => p.required)) {
          const field = block.fields.find(
            (f) => f.in === "body" && (f.target ?? f.name) === param.name
          );
          expect(field?.required, `${block.type} requires ${param.name}`).toBe(
            true
          );
        }
      }
    }
  });
});
