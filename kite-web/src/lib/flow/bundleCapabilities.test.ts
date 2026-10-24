import { describe, expect, it } from "vitest";
import { blockDefinitions } from "../blocks";
import {
  bundleCapabilities,
  capabilityGroups,
  harmlessTypes,
} from "./bundleCapabilities";

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
      bundleCapabilities([
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
      ])
    ).toEqual([
      "Bans, unbans, kicks or times out members",
      "Uses ER:LC with your credentials",
      "Reads your secret API_KEY",
      "Has IDs of channels, roles or users of the server it was made in",
    ]);
  });
});
