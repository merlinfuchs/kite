import { describe, expect, it } from "vitest";
import { testEdge, testNode } from "./testUtils";
import { getAvailablePlaceholders } from "./placeholders";

const node = testNode;
const edge = testEdge;

const nodes = [
  node("entry", "entry_command"),
  node("arg", "option_command_argument", { name: "user" }),
  node("get", "action_user_get", { temporary_name: "target" }),
  node("msg", "action_response_create", {
    message_data: { components: [{ type: 1, components: [{ id: 7 }] }] },
  }),
  node("log", "action_log", { custom_label: "Logger" }),
];

const edges = [
  edge("arg", "entry"),
  edge("entry", "get"),
  edge("get", "msg"),
  edge("msg", "log", "component_7"),
];

const values = (nodeId?: string) =>
  getAvailablePlaceholders(nodeId, nodes, edges, "command").map((g) => [
    g.label,
    g.placeholders.map((p) => p.value),
  ]);

describe("getAvailablePlaceholders", () => {
  it("only lists flow wide placeholders without a block", () => {
    expect(values().map(([label]) => label)).toEqual([
      "Command",
      "User",
      "Server",
      "Channel",
      "App",
    ]);
    expect(values()[0]).toEqual(["Command", ["arg('user')"]]);
  });

  it("lists the server and member placeholders", () => {
    const groups = Object.fromEntries(values());
    expect(groups["Server"]).toEqual([
      "guild.id",
      "guild.name",
      "guild.icon_url",
      "guild.member_count",
      "guild.boost_count",
      "guild.owner_id",
      "guild.boost_level",
      "guild.created_at",
      "guild.banner_url",
      "guild.description",
      "guild.vanity_url",
      "guild.role_count",
      "guild.channel_count",
      "guild.emoji_count",
      "guild.rules_channel",
      "guild.system_channel",
    ]);
    expect(groups["Channel"]).toEqual([
      "channel.id",
      "channel.name",
      "channel.mention",
      "channel.type",
      "channel.category_id",
      "channel.category_name",
    ]);
    expect(groups["User"]).toEqual(
      expect.arrayContaining([
        "user.is_bot",
        "user.created_at",
        "user.joined_at",
        "user.top_role",
        "user.role_mentions",
        "user.role_names",
        "user.role_count",
        "user.color",
        "user.is_booster",
        "user.boosting_since",
        "user.is_timed_out",
        "user.timeout_until",
        "user.is_owner",
        "user.is_admin",
        "user.permissions",
      ])
    );
  });

  it("adds the results and variables of earlier blocks", () => {
    expect(values("msg").slice(5)).toEqual([
      ["Temporary Variables", ["var('target')"]],
      ["Node Results", ["result('get')"]],
    ]);
  });

  it("adds the original interaction after a resume point", () => {
    const labels = values("log").map(([label]) => label);
    expect(labels).toContain("Original User");
    expect(labels).not.toContain("Previous User");
    expect(values("log").at(-1)).toEqual([
      "Node Results",
      ["result('msg')", "result('get')"],
    ]);
  });

  it("lists all picked options of modal inputs in the modal's sub-flow", () => {
    const input = (custom_id: string, type: string, max_values?: number) => ({
      type: "label",
      label: custom_id,
      components: [
        {
          type,
          custom_id,
          max_values,
          options: [{ label: "A" }, { label: "B" }],
        },
      ],
    });
    const modalNodes = [
      node("entry", "entry_command"),
      node("modal", "suspend_response_modal", {
        modal_data: {
          title: "Form",
          components: [
            input("single", "string_select"),
            input("multi", "string_select", 2),
            input("roles", "role_select", 3),
            input("extras", "checkbox_group"),
            input("size", "radio_group"),
          ],
        },
      }),
      node("log", "action_log"),
      node("msg", "action_response_create", {
        message_data: { components: [{ type: 1, components: [{ id: 7 }] }] },
      }),
      node("later", "action_log"),
    ];
    const modalEdges = [
      edge("entry", "modal"),
      edge("modal", "log"),
      edge("log", "msg"),
      edge("msg", "later", "component_7"),
    ];
    const modalInputs = (nodeId: string) =>
      getAvailablePlaceholders(nodeId, modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.map((p) => p.value);

    const expected = [
      "input('single')",
      "input('multi')",
      "inputs('multi')",
      "input('roles')",
      "inputs('roles')",
      "input('extras')",
      "inputs('extras')",
      "input('size')",
    ];
    expect(modalInputs("log")).toEqual(expected);
    expect(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.find((p) => p.value === "inputs('extras')")?.label
    ).toBe("extras (Selected Values)");
    // Like input(), inputs() still works after the next resume point.
    expect(modalInputs("later")).toEqual(expected);
  });

  it("lists the reaction emoji for discord events", () => {
    const groups = getAvailablePlaceholders(undefined, [], [], "event_discord");
    expect(groups.find((g) => g.label === "Emoji")?.placeholders).toEqual([
      { label: "Emoji", value: "emoji" },
      { label: "Emoji ID", value: "emoji.id" },
      { label: "Emoji Name", value: "emoji.name" },
      { label: "Emoji Mention", value: "emoji.mention" },
    ]);
  });

  it("lists forum post placeholders in their own category", () => {
    const groups = getAvailablePlaceholders(undefined, [], [], "event_discord");
    const forumPost = groups.find((g) => g.label === "Forum Post");

    expect(forumPost?.placeholders).toEqual([
      { label: "Forum Post", value: "forum_post" },
      { label: "Post ID", value: "forum_post.id" },
      { label: "Post Title", value: "forum_post.title" },
      { label: "Post URL", value: "forum_post.url" },
      { label: "Post Author ID", value: "forum_post.author_id" },
      {
        label: "Parent Forum Channel ID",
        value: "forum_post.parent_channel_id",
      },
      { label: "Applied Tag IDs", value: "forum_post.tag_ids" },
      { label: "Applied Tag Names", value: "forum_post.tags" },
      { label: "Post Created At (Unix)", value: "forum_post.created_at" },
      { label: "Post Message Count", value: "forum_post.message_count" },
      { label: "Post Member Count", value: "forum_post.member_count" },
      { label: "Post Archived", value: "forum_post.archived" },
      { label: "Post Locked", value: "forum_post.locked" },
      {
        label: "Post Auto-Archive Duration (Minutes)",
        value: "forum_post.auto_archive_duration",
      },
      { label: "Forum Channel", value: "forum" },
    ]);
  });
});
