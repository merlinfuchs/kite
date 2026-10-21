import { describe, expect, it } from "vitest";
import { testEdge, testNode } from "./testUtils";
import { flattenPlaceholders, getAvailablePlaceholders } from "./placeholders";

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

// The values with the fields of each value.
const allValues = (nodeId?: string) =>
  getAvailablePlaceholders(nodeId, nodes, edges, "command").map((g) => [
    g.label,
    flattenPlaceholders(g.placeholders).map((p) => p.value),
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
    const groups = Object.fromEntries(allValues());
    expect(groups["Server"]).toEqual([
      "guild",
      "guild.id",
      "guild.name",
      "guild.icon_url",
      "guild.owner_id",
      "guild.member_count",
      "guild.boost_count",
      "guild.boost_level",
      "guild.created_at",
      "guild.banner_url",
      "guild.description",
      "guild.vanity_url",
      "guild.role_count",
      "guild.channel_count",
      "guild.emoji_count",
      "guild.rules_channel",
      "guild.rules_channel.id",
      "guild.rules_channel.name",
      "guild.rules_channel.mention",
      "guild.rules_channel.type",
      "guild.rules_channel.category_id",
      "guild.rules_channel.category_name",
      "guild.system_channel",
      "guild.system_channel.id",
      "guild.system_channel.name",
      "guild.system_channel.mention",
      "guild.system_channel.type",
      "guild.system_channel.category_id",
      "guild.system_channel.category_name",
    ]);
    expect(groups["Channel"]).toEqual([
      "channel",
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
        "user.top_role.name",
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
    const [variables, results] = values("msg").slice(5);
    expect(variables).toEqual(["Temporary Variables", ["var('target')"]]);
    expect(results[0]).toBe("Node Results");
    expect(results[1]).toEqual(["result('get')"]);
    // Get User returns a user, so its fields are listed.
    expect(allValues("msg").at(-1)![1]).toContain("result('get').display_name");
  });

  it("leaves out member fields a block result doesn't know", () => {
    const resultNodes = [
      node("entry", "entry_command"),
      node("get", "action_member_get"),
      node("log", "action_log"),
    ];
    const resultEdges = [edge("entry", "get"), edge("get", "log")];
    const results = flattenPlaceholders(
      getAvailablePlaceholders("log", resultNodes, resultEdges, "command").find(
        (g) => g.label === "Node Results"
      )!.placeholders
    ).map((p) => p.value);

    expect(results).toContain("result('get').nick");
    expect(results).not.toContain("result('get').is_admin");
    expect(results).not.toContain("result('get').top_role");
    // The interaction knows the server of its user.
    expect(Object.fromEntries(allValues())["User"]).toContain("user.is_admin");
  });

  it("lists the fields of command arguments by their type", () => {
    const argNodes = [
      node("entry", "entry_command"),
      node("target", "option_command_argument", {
        name: "target",
        command_argument_type: "user",
      }),
      node("reason", "option_command_argument", {
        name: "reason",
        command_argument_type: "string",
      }),
    ];
    const argEdges = [edge("target", "entry"), edge("reason", "entry")];
    const command = getAvailablePlaceholders(
      undefined,
      argNodes,
      argEdges,
      "command"
    ).find((g) => g.label === "Command")!.placeholders;
    const flat = flattenPlaceholders(command);

    // The fields are listed under the argument.
    expect(command.map((p) => p.value)).toEqual([
      "arg('target')",
      "arg('reason')",
    ]);
    expect(flat.map((p) => p.value)).toContain("arg('target').top_role.name");
    expect(flat.filter((p) => p.value.startsWith("arg('reason')"))).toEqual([
      { label: "Command Arg 'reason'", value: "arg('reason')" },
    ]);
    expect(flat.find((p) => p.value === "arg('target').avatar_url")).toEqual({
      label: "Command Arg 'target' Avatar URL",
      value: "arg('target').avatar_url",
    });
  });

  it("adds the original interaction after a resume point", () => {
    const labels = values("log").map(([label]) => label);
    expect(labels).toContain("Original User");
    expect(labels).not.toContain("Previous User");
    const [label, results] = values("log").at(-1)!;
    expect(label).toBe("Node Results");
    expect(results).toEqual(["result('msg')", "result('get')"]);
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
            input("image", "file_upload"),
            input("images", "file_upload", 10),
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
        .find((g) => g.label === "Modal Inputs")!
        .placeholders.map((p) => p.value);

    const expected = [
      "input('single')",
      "input('multi')",
      "inputs('multi')",
      "input('roles')",
      "inputs('roles')",
      "picked('roles')",
      "pickedAll('roles')",
      "input('extras')",
      "inputs('extras')",
      "input('size')",
      "input('image')",
      "picked('image')",
      "input('images')",
      "inputs('images')",
      "picked('images')",
      "pickedAll('images')",
    ];
    expect(modalInputs("log")).toEqual(expected);
    expect(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.find((p) => p.value === "inputs('extras')")?.label
    ).toBe("extras (Selected Values)");
    expect(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.find((p) => p.value === "pickedAll('images')")?.label
    ).toBe("images (All Files)");
    expect(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command")
        .find((g) => g.label === "Modal Inputs")
        ?.placeholders.find((p) => p.value === "pickedAll('roles')")?.label
    ).toBe("roles (All Roles)");
    const fields = flattenPlaceholders(
      getAvailablePlaceholders("log", modalNodes, modalEdges, "command").find(
        (g) => g.label === "Modal Inputs"
      )!.placeholders
    ).map((p) => p.value);
    expect(fields).toContain("picked('roles').color");
    expect(fields).toContain("join(map(pickedAll('roles'), #.name), ', ')");
    expect(fields).toContain("picked('image').url");
    // join only takes strings, so the sizes of all files aren't listed.
    expect(fields).toContain("join(map(pickedAll('images'), #.url), ', ')");
    expect(fields).not.toContain(
      "join(map(pickedAll('images'), #.size), ', ')"
    );
    // Like input(), inputs() still works after the next resume point.
    expect(modalInputs("later")).toEqual(expected);
  });

  it("lists the reaction emoji for discord events", () => {
    const groups = getAvailablePlaceholders(undefined, [], [], "event_discord");
    expect(groups.find((g) => g.label === "Emoji")?.placeholders).toEqual([
      { label: "Emoji", value: "emoji" },
      {
        label: "Emoji ID",
        value: "emoji.id",
        description: "Empty for a unicode emoji",
      },
      { label: "Emoji Name", value: "emoji.name" },
      { label: "Emoji Animated", value: "emoji.animated" },
      {
        label: "Emoji Mention",
        value: "emoji.mention",
        description:
          "The emoji the way Discord shows it, like 🔥 or <:name:id>",
      },
    ]);
  });
});
