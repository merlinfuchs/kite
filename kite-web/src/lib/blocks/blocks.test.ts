import { describe, expect, it } from "vitest";
import { getIntegration, integrations } from "../integrations";
import { integrationApis } from "../integrations/apis";
import {
  blockDefinitions,
  blockIntegrations,
  requestBlocks,
  requestChoiceBlocks,
  sendsRequest,
} from ".";
import { getNodeValues } from "../flow/nodes";

// What the service needs: how each block runs, which integrations it needs
// and the fields of request blocks. Schemas are only for the editor and the
// flow AI, and credits that depend on the settings are computed in Go.
const serviceBlocks = blockDefinitions.map((block) => ({
  type: block.type,
  credits: typeof block.credits === "number" ? block.credits : null,
  audit_log_reason: !!block.audit_log_reason,
  requires: blockIntegrations(block),
  // Labels of operations are only for the editor.
  run:
    block.run.kind === "request_choice"
      ? {
          ...block.run,
          setting_label: undefined,
          operations: block.run.operations.map((o) => ({
            id: o.id,
            method: o.method,
            path: o.path,
            fields: o.fields.map((f) => ({
              name: f.name,
              in: f.in,
              target: f.target ?? f.name,
              required: !!f.required,
            })),
          })),
        }
      : block.run,
  fields: (block.run.kind === "custom" ? [] : block.fields ?? []).map((f) => ({
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
      JSON.stringify({ integrations, blocks: serviceBlocks }, null, 2) + "\n"
    ).toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/block_definitions.json"
    );
  });

  it("have short ID prefixes", () => {
    for (const block of blockDefinitions) {
      expect(block.id_prefix, block.type).toMatch(/^[a-z]{2,12}$/);
    }
  });

  it("have unique types and field names", () => {
    const types = blockDefinitions.map((b) => b.type);
    expect(new Set(types).size).toBe(types.length);
    for (const block of blockDefinitions) {
      const names = (block.fields ?? []).map((f) => f.name);
      expect(new Set(names).size, block.type).toBe(names.length);
    }
  });

  it("send requests with fields and a fixed cost", () => {
    for (const block of blockDefinitions.filter(sendsRequest)) {
      expect(block.fields, block.type).toBeDefined();
      expect(typeof block.credits, block.type).toBe("number");
    }
  });

  it("describe the fields without a schema", () => {
    for (const block of blockDefinitions) {
      for (const field of (block.fields ?? []).filter((f) => !f.schema)) {
        expect(field.label, `${block.type}.${field.name}`).toBeTruthy();
        expect(field.description, `${block.type}.${field.name}`).toBeTruthy();
      }
    }
  });

  // Generated schemas store every value as text, which only requests convert.
  // Go reads the settings of custom blocks into typed fields.
  it("give the fields of custom blocks a schema", () => {
    for (const block of blockDefinitions.filter(
      (b) => b.run.kind === "custom"
    )) {
      for (const field of block.fields ?? []) {
        expect(field.schema, `${block.type}.${field.name}`).toBeDefined();
      }
    }
  });

  it("type the fields of requests and those without a schema", () => {
    for (const block of blockDefinitions) {
      for (const field of block.fields ?? []) {
        if (block.run.kind === "request" || !field.schema) {
          expect(field.type, `${block.type}.${field.name}`).toBeDefined();
        }
      }
    }
  });

  it("only send the fields of request blocks", () => {
    for (const block of blockDefinitions) {
      for (const field of block.fields ?? []) {
        expect(!!field.in, `${block.type}.${field.name}`).toBe(
          block.run.kind === "request"
        );
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

  it("pick operations that send their own fields", () => {
    for (const block of requestChoiceBlocks()) {
      const ids = block.run.operations.map((o) => o.id);
      expect(new Set(ids).size, block.type).toBe(ids.length);
      expect(
        block.fields.some((f) => f.name === block.run.setting),
        `${block.type} has a field named like its setting`
      ).toBe(false);
      for (const op of block.run.operations) {
        expect(op.label, `${block.type} ${op.id}`).toBeTruthy();
        expect(op.description, `${block.type} ${op.id}`).toBeTruthy();
        const names = op.fields.map((f) => f.name);
        expect(new Set(names).size, `${block.type} ${op.id}`).toBe(
          names.length
        );
        for (const name of names) {
          expect(
            block.fields.some((f) => f.name === name),
            `${block.type} ${op.id} ${name}`
          ).toBe(true);
        }
      }
    }
  });

  // The URL would contain the credential.
  it("only return URLs of integrations without a credential", () => {
    for (const block of requestChoiceBlocks().filter(
      (b) => b.run.response === "url"
    )) {
      expect(getIntegration(block.run.integration)?.auth.type).toBe("none");
    }
  });

  it("pick operations of their integration's API", () => {
    for (const block of requestChoiceBlocks()) {
      const api = integrationApis[block.run.integration];
      for (const op of block.run.operations) {
        const spec = api?.operations.find((o) => o.id === op.id);
        const name = `${block.type} ${op.id}`;
        expect(spec, name).toBeDefined();
        if (!spec) continue;

        expect([op.method, op.path], name).toEqual([spec.method, spec.path]);
        const targets = (location: string) =>
          op.fields
            .filter((f) => f.in === location)
            .map((f) => f.target ?? f.name)
            .sort();
        expect(targets("path"), name).toEqual(
          spec.path_params.map((p) => p.name).sort()
        );
        for (const target of targets("query")) {
          expect(
            spec.query_params.some((p) => p.name === target),
            `${name}.${target}`
          ).toBe(true);
        }
        for (const param of spec.query_params.filter((p) => p.required)) {
          const field = op.fields.find(
            (f) => f.in === "query" && (f.target ?? f.name) === param.name
          );
          expect(field?.required, `${name} requires ${param.name}`).toBe(true);
        }
      }
    }
  });

  it("match their integration's API", () => {
    for (const block of requestBlocks()) {
      const op = integrationApis[block.run.integration]?.operations.find(
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

// Blocks that had a hand-written schema, which ignored settings it didn't
// know, so saved flows can have some. Every other block rejects them, so a
// misnamed setting isn't lost. Remove blocks from here, never add them.
const lenientBlocks = new Set([
  "action_ai_chat_completion",
  "action_ai_web_search",
  "action_channel_create",
  "action_channel_delete",
  "action_channel_edit",
  "action_channel_get",
  "action_discord_api_request",
  "action_expression_evaluate",
  "action_guild_get",
  "action_http_request",
  "action_log",
  "action_member_ban",
  "action_member_edit",
  "action_member_get",
  "action_member_kick",
  "action_member_role_add",
  "action_member_role_remove",
  "action_member_timeout",
  "action_member_unban",
  "action_message_create",
  "action_message_delete",
  "action_message_edit",
  "action_message_get",
  "action_message_pin",
  "action_message_reaction_create",
  "action_message_reaction_delete",
  "action_message_unpin",
  "action_poll_create",
  "action_private_message_create",
  "action_random_generate",
  "action_response_create",
  "action_response_defer",
  "action_response_delete",
  "action_response_edit",
  "action_roblox_user_get",
  "action_role_get",
  "action_status_set",
  "action_thread_create",
  "action_thread_member_add",
  "action_thread_member_remove",
  "action_user_get",
  "action_variable_delete",
  "action_variable_get",
  "action_variable_set",
  "action_voice_channel_join",
  "action_voice_channel_leave",
  "control_condition_channel",
  "control_condition_compare",
  "control_condition_item_channel",
  "control_condition_item_compare",
  "control_condition_item_else",
  "control_condition_item_role",
  "control_condition_item_user",
  "control_condition_role",
  "control_condition_user",
  "control_error_handler",
  "control_loop",
  "control_loop_each",
  "control_loop_end",
  "control_loop_exit",
  "control_sleep",
  "entry_command",
  "entry_component_button",
  "entry_event",
  "option_command_argument",
  "option_command_contexts",
  "option_command_permissions",
  "option_event_filter",
  "suspend_response_modal",
]);

describe("unknown settings", () => {
  it("are rejected by new blocks", () => {
    for (const block of blockDefinitions) {
      expect(
        !!block.strict_settings || lenientBlocks.has(block.type),
        `${block.type} should set strict_settings`
      ).toBe(true);
    }
  });
});

describe("Respond to webhook", () => {
  const schema = getNodeValues("action_webhook_response_create").dataSchema!;

  it("only allows the statuses the service does", () => {
    for (const [status, valid] of [
      ["200", true],
      ["401", true],
      ["599", true],
      ["{{ var('status') }}", true],
      ["302", false],
      ["100", false],
      ["600", false],
      ["", false],
      ["2O0", false],
    ] as const) {
      expect(
        schema.safeParse({ webhook_response_status: status }).success,
        status
      ).toBe(valid);
    }
  });
});

describe("Generate random value", () => {
  const schema = getNodeValues("action_random_generate").dataSchema!;

  it("requires the settings of its mode", () => {
    for (const [data, valid] of [
      // Blocks from before the item mode have no mode.
      [{ random_min: "1", random_max: "10" }, true],
      [{ random_mode: "number", random_min: "1", random_max: "10" }, true],
      [{ random_mode: "number", random_items: "a" }, false],
      [{ random_mode: "item", random_items: "a\nb" }, true],
      [{ random_mode: "item", random_items: " " }, false],
      [{ random_mode: "item", random_items: "a".repeat(10001) }, false],
      [{ random_mode: "dice", random_items: "a" }, false],
    ] as const) {
      expect(schema.safeParse(data).success, JSON.stringify(data)).toBe(valid);
    }
  });
});
