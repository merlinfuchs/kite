import { describe, expect, it } from "vitest";
import { getIntegration, integrations } from "../integrations";
import { integrationApis } from "../integrations/apis";
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
  fields: (block.run.kind === "request" ? block.fields ?? [] : []).map((f) => ({
    name: f.name,
    in: f.in,
    target: f.target ?? f.name,
    type: f.type,
    required: !!f.required,
    fallback: f.fallback ?? "",
    min: f.min ?? null,
    max: f.max ?? null,
    max_length: f.max_length ?? null,
    options: f.options?.map((o) => o.value) ?? null,
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

  it("have unique types and field names", () => {
    const types = blockDefinitions.map((b) => b.type);
    expect(new Set(types).size).toBe(types.length);
    for (const block of blockDefinitions) {
      const names = (block.fields ?? []).map((f) => f.name);
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

  it("only send the bot token to integrations, with Discord required", () => {
    for (const block of requestBlocks().filter((b) =>
      b.run.inject?.some((i) => i.value === "discord_bot_token")
    )) {
      expect(block.run.integration, block.type).not.toBe("discord");
      expect(block.requires, block.type).toContain("discord");
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
      for (const param of op.query_params.filter((p) => p.required)) {
        const field = block.fields.find(
          (f) => f.in === "query" && (f.target ?? f.name) === param.name
        );
        expect(
          field?.required || !!field?.fallback,
          `${block.type} requires ${param.name}`
        ).toBe(true);
      }
      for (const field of block.fields.filter((f) => f.options)) {
        const param = [...op.query_params, ...(op.body_params ?? [])].find(
          (p) => p.name === (field.target ?? field.name)
        );
        expect(param?.enum, `${block.type}.${field.name}`).toEqual(
          expect.arrayContaining(field.options!.map((o) => o.value))
        );
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
          const injected = block.run.inject?.some(
            (i) => i.in === "body" && i.name === param.name
          );
          expect(
            field?.required || injected,
            `${block.type} requires ${param.name}`
          ).toBe(true);
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
  "action_forum_post_create",
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
