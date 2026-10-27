import { describe, expect, it } from "vitest";
import { blockDefinitions } from "../blocks";
import {
  bundleCapabilities,
  capabilityGroups,
  harmlessTypes,
} from "./bundleCapabilities";
import { noteNodeType } from "./notes";

describe("bundleCapabilities", () => {
  // A new block has to be pointed out on bundle pages or be harmless.
  it("knows every action and control block", () => {
    const known = new Set([
      ...capabilityGroups.flatMap((g) => g.types),
      ...harmlessTypes,
    ]);
    const unknown = blockDefinitions
      .map((b) => b.type)
      .filter(
        (type) =>
          (type.startsWith("action_") || type.startsWith("control_")) &&
          !known.has(type)
      );
    expect(unknown).toEqual([]);
  });

  it("lists what the blocks do and the integrations they use", () => {
    expect(
      bundleCapabilities({
        flows: [
          {
            nodes: [
              { type: "action_member_ban" },
              { type: "action_erlc_command_run" },
              {
                type: "action_message_create",
                data: {
                  message_data: {
                    content: "{{secrets.API_KEY}} <#123456789012345678>",
                  },
                },
              },
            ],
          },
        ],
      })
    ).toEqual([
      "Bans, unbans, kicks or times out members",
      "Uses ER:LC with your credentials",
      "Reads your secret API_KEY",
      "Has IDs of channels, roles or users of the server it was made in",
    ]);
  });

  const keyWarning =
    "Seems to have an API key or token typed in, instead of using a secret";

  const request = (headers: { key: string; value: string }[], url = "") => [
    {
      nodes: [
        {
          type: "action_http_request",
          data: { http_request_data: { url, method: "GET", headers } },
        },
      ],
    },
  ];

  it("warns about keys typed into blocks", () => {
    for (const flows of [
      request([{ key: "Authorization", value: "Bearer abcdefghijklmnop1234" }]),
      request([{ key: "X-API-Key", value: "a1b2c3d4e5f6g7h8i9j0" }]),
      request([
        { key: "X-Custom", value: "sk-or-v1-abcdefghijklmnopqrstuvwxyz" },
      ]),
      request([], "https://api.example.com/v1?api_key=a1b2c3d4e5f6g7h8i9j0"),
    ]) {
      expect(bundleCapabilities({ flows })).toContain(keyWarning);
    }
  });

  it("warns about keys in shared defaults", () => {
    expect(
      bundleCapabilities({
        flows: [],
        defaults: ["sk-proj-abcdefghijklmnopqrstuvwxyz"],
      })
    ).toEqual([keyWarning]);
  });

  it("warns about keys in notes and message templates", () => {
    expect(
      bundleCapabilities({
        flows: [
          {
            nodes: [
              {
                type: noteNodeType,
                data: { text: "key: sk-proj-abcdefghijklmnopqrstuvwxyz" },
              },
            ],
          },
        ],
      })
    ).toEqual([keyWarning]);
    expect(
      bundleCapabilities({
        flows: [],
        messages: [{ content: "ghp_abcdefghijklmnopqrstuvwxyz0123456789" }],
      })
    ).toEqual([keyWarning]);
  });

  it("treats webhook URLs as keys, not IDs", () => {
    const url = `https://discord.com/api/webhooks/123456789012345678/${"a".repeat(
      68
    )}`;
    expect(bundleCapabilities({ flows: request([], url) })).toEqual([
      "Sends requests to other websites",
      keyWarning,
    ]);
  });

  it("warns about IDs in message templates", () => {
    expect(
      bundleCapabilities({
        flows: [],
        messages: [{ content: "<#123456789012345678>" }],
      })
    ).toEqual([
      "Has IDs of channels, roles or users of the server it was made in",
    ]);
  });

  it("doesn't warn about secrets and placeholders", () => {
    expect(
      bundleCapabilities({
        flows: request(
          [
            { key: "Authorization", value: "Bearer {{secrets.API_KEY}}" },
            { key: "X-API-Key", value: "{{params.key}}" },
            { key: "Accept", value: "application/json" },
          ],
          "https://api.example.com/v1?api_key={{secrets.API_KEY}}"
        ),
      })
    ).not.toContain(keyWarning);
  });
});
