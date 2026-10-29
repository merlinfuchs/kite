import { describe, expect, it } from "vitest";
import {
  flowPermissions,
  missingPermissions,
  templatePermissionFaq,
  templatePermissions,
} from "./templatePermissions";
import { Template } from "./templates";
import {
  encodePermissionsBitset,
  permissionBits,
} from "../discord/permissions";

describe("templatePermissions", () => {
  it("derives permissions from flows", () => {
    const permissions = flowPermissions({
      nodes: [
        {
          id: "1",
          type: "action_member_ban",
          data: {},
          position: { x: 0, y: 0 },
        },
        {
          id: "2",
          type: "action_message_create",
          data: {},
          position: { x: 0, y: 0 },
        },
        { id: "3", type: "note", data: {}, position: { x: 0, y: 0 } },
      ],
      edges: [],
    });
    expect(permissions).toEqual(
      expect.arrayContaining(["Ban Members", "View Channel", "Send Messages"])
    );
  });

  it("detects role hierarchy for moderation and role blocks", () => {
    const banOnly = templatePermissions({
      commands: [
        {
          key: "ban",
          name: "ban",
          description: "",
          parameterKeys: [],
          flowSource: {
            nodes: [
              {
                id: "1",
                type: "action_member_ban",
                data: {},
                position: { x: 0, y: 0 },
              },
            ],
            edges: [],
          },
        },
      ],
      eventListeners: [],
      messages: [],
    } as unknown as Template);
    expect(banOnly.needsRoleHierarchy).toBe(true);

    const messageOnly = templatePermissions({
      commands: [
        {
          key: "msg",
          name: "msg",
          description: "",
          parameterKeys: [],
          flowSource: {
            nodes: [
              {
                id: "1",
                type: "action_message_create",
                data: {},
                position: { x: 0, y: 0 },
              },
            ],
            edges: [],
          },
        },
      ],
      eventListeners: [],
      messages: [],
    } as unknown as Template);
    expect(messageOnly.needsRoleHierarchy).toBe(false);
  });

  it("derives permissions from messages posted on install", () => {
    const template = {
      commands: [],
      eventListeners: [],
      messages: [
        {
          key: "welcome",
          name: "Welcome",
          description: null,
          parameterKeys: [],
          sendTo: "welcome_channel",
          customize: false,
          data: {
            embeds: [{ title: "Welcome" }],
          },
          flowSources: {},
        },
      ],
    } as unknown as Template;

    const res = templatePermissions(template);
    expect(res.permissions).toEqual(
      expect.arrayContaining(["View Channel", "Send Messages", "Embed Links"])
    );
  });

  it("handles empty flow and template", () => {
    expect(flowPermissions({ nodes: [], edges: [] })).toEqual([]);

    const emptyTemplate = templatePermissions({
      commands: [],
      eventListeners: [],
      messages: [],
    } as unknown as Template);
    expect(emptyTemplate.permissions).toEqual([]);
    expect(emptyTemplate.needsRoleHierarchy).toBe(false);
  });

  it("filters permissions using includedKeys subset", () => {
    const template = {
      commands: [
        {
          key: "cmd_ban",
          name: "ban",
          description: "",
          parameterKeys: [],
          flowSource: {
            nodes: [
              {
                id: "1",
                type: "action_member_ban",
                data: {},
                position: { x: 0, y: 0 },
              },
            ],
            edges: [],
          },
        },
        {
          key: "cmd_kick",
          name: "kick",
          description: "",
          parameterKeys: [],
          flowSource: {
            nodes: [
              {
                id: "2",
                type: "action_member_kick",
                data: {},
                position: { x: 0, y: 0 },
              },
            ],
            edges: [],
          },
        },
      ],
      eventListeners: [],
      messages: [],
    } as unknown as Template;

    // Only include cmd_kick
    const derived = templatePermissions(template, new Set(["cmd_kick"]));
    expect(derived.permissions).toEqual(["Kick Members"]);
    expect(derived.needsRoleHierarchy).toBe(true);
  });

  it("detects attachments in messages and message create blocks", () => {
    const template = {
      commands: [
        {
          key: "post",
          name: "post",
          description: "",
          parameterKeys: [],
          flowSource: {
            nodes: [
              {
                id: "1",
                type: "action_message_create",
                data: {
                  message_data: {
                    content: "hello",
                    embeds: [{ title: "embed" }],
                    attachments: [{ filename: "image.png" }],
                  },
                },
                position: { x: 0, y: 0 },
              },
            ],
            edges: [],
          },
        },
      ],
      eventListeners: [],
      messages: [
        {
          key: "welc_msg",
          name: "Welcome",
          sendTo: "welcome_channel",
          data: {
            content: "hi",
            attachments: [{ filename: "banner.png" }],
          },
        },
      ],
    } as unknown as Template;

    const derived = templatePermissions(template);
    expect(derived.permissions).toContain("Embed Links");
    expect(derived.permissions).toContain("Attach Files");
  });

  describe("missingPermissions", () => {
    it("handles null, undefined, 0, and all permissions bitmask", () => {
      expect(missingPermissions(["Ban Members"], null)).toEqual([]);
      expect(missingPermissions(["Ban Members"], undefined)).toEqual([]);
      expect(missingPermissions(["Ban Members"], "0")).toEqual(["Ban Members"]);

      // ^discord.Permissions(0) in Go wire format
      expect(
        missingPermissions(
          ["Ban Members", "Moderate Members"],
          "18446744073709551615"
        )
      ).toEqual([]);
    });
    it("returns empty when bot has Administrator", () => {
      const adminBit = permissionBits.find(
        (p) => p.label === "Administrator"
      )!.bit;
      const bitset = encodePermissionsBitset([adminBit]);
      const missing = missingPermissions(
        ["Ban Members", "Manage Roles"],
        bitset
      );
      expect(missing).toEqual([]);
    });

    it("identifies missing permissions against the bot's roles", () => {
      const viewBit = permissionBits.find(
        (p) => p.label === "View Channel"
      )!.bit;
      const sendBit = permissionBits.find(
        (p) => p.label === "Send Messages"
      )!.bit;
      const bitset = encodePermissionsBitset([viewBit, sendBit]);

      const missing = missingPermissions(
        ["View Channel", "Send Messages", "Ban Members"],
        bitset
      );
      expect(missing).toEqual(["Ban Members"]);
    });

    it("returns empty when bot has all required permissions", () => {
      const viewBit = permissionBits.find(
        (p) => p.label === "View Channel"
      )!.bit;
      const sendBit = permissionBits.find(
        (p) => p.label === "Send Messages"
      )!.bit;
      const bitset = encodePermissionsBitset([viewBit, sendBit]);

      const missing = missingPermissions(
        ["View Channel", "Send Messages"],
        bitset
      );
      expect(missing).toEqual([]);
    });
  });

  describe("templatePermissionFaq", () => {
    it("generates a natural FAQ entry for permissions and hierarchy", () => {
      const template = {
        commands: [
          {
            key: "ban",
            name: "ban",
            description: "",
            parameterKeys: [],
            flowSource: {
              nodes: [
                {
                  id: "1",
                  type: "action_member_ban",
                  data: {},
                  position: { x: 0, y: 0 },
                },
                {
                  id: "2",
                  type: "action_message_create",
                  data: {},
                  position: { x: 0, y: 0 },
                },
              ],
              edges: [],
            },
          },
        ],
        eventListeners: [],
        messages: [],
      } as unknown as Template;

      const faq = templatePermissionFaq(template);
      expect(faq?.question).toBe("What permissions does the bot need?");
      expect(faq?.answer).toContain("Ban Members");
      expect(faq?.answer).toContain("Send Messages");
      expect(faq?.answer).toContain("Keep the bot's role above the members");
    });

    it("returns null when no permissions or role hierarchy are needed", () => {
      const template = {
        commands: [],
        eventListeners: [],
        messages: [],
      } as unknown as Template;

      expect(templatePermissionFaq(template)).toBeNull();
    });
  });
});
