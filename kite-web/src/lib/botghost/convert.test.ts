import { describe, expect, it } from "vitest";
import LZString from "lz-string";
import { convertBotGhostItem } from "./convert";
import { parseBotGhostExport } from "./parse";

const everyone = {
  allowed_roles: [{ name: "@everyone", id: "everyone" }],
  banned_roles: [],
  banned_channels: [],
  banned_users: [],
  required_permissions: [],
};

function convert(data: unknown) {
  const parsed = parseBotGhostExport(JSON.stringify(data));
  return [...parsed.commands, ...parsed.events].map(convertBotGhostItem);
}

function blocks(item: ReturnType<typeof convert>[number]) {
  return item.flow!.nodes.filter((n) => n.type !== "note");
}

function find(item: ReturnType<typeof convert>[number], type: string) {
  return blocks(item).filter((n) => n.type === type);
}

// The block that runs after source, by the output of source.
function after(
  item: ReturnType<typeof convert>[number],
  source: string,
  handle?: string
) {
  const edge = item.flow!.edges.find(
    (e) => e.source === source && (!handle || e.sourceHandle === handle)
  );
  return item.flow!.nodes.find((n) => n.id === edge?.target);
}

describe("convertBotGhostItem", () => {
  it("converts replies, options and variables", () => {
    const [item] = convert({
      name: "Suggest",
      description: "Send a suggestion",
      options: [
        {
          type: "channel",
          name: "channel",
          description: "Where to post",
          required: true,
        },
        { type: "string", name: "your-idea", description: "", required: true },
      ],
      actions: [
        {
          type: "targeted_response",
          response_options: {
            embed: {
              title: "New suggestion",
              description: "{option_your-idea}\n\nBy {user} in {server}",
              color: "#08f70b",
              footer: "Vote below",
              thumbnail: "{server_icon}",
            },
            emojis: ["👍", "<:no:123456789012345678>"],
          },
          target: { variable: "channel" },
        },
        { type: "plain_text", response: "Sent to {option_channel}!" },
      ],
      permissions: { ...everyone, required_permissions: ["MANAGE_MESSAGES"] },
      cooldown: { type: "user", interval: 2, interval_type: "minutes" },
      ephemeral: true,
    });

    expect(item.status).toBe("ready");
    expect(item.notes).toEqual([]);
    expect(item.name).toBe("suggest");

    const args = find(item, "option_command_argument").map((n) => n.data);
    expect(args.map((a) => a.name)).toEqual(["channel", "your_idea"]);
    expect(args[1].description).toBe("your_idea");
    expect(find(item, "option_command_permissions")[0].data).toEqual({
      command_permissions: "8192",
    });
    expect(find(item, "option_command_cooldown")[0].data).toMatchObject({
      cooldown_scope: "user",
      cooldown_duration_seconds: "120",
    });

    const [message] = find(item, "action_message_create");
    expect(message.data.channel_target).toBe("{{arg('channel').id}}");
    expect(message.data.message_data).toEqual({
      embeds: [
        {
          title: "New suggestion",
          description:
            "{{arg('your_idea')}}\n\nBy {{user.mention}} in {{guild.name}}",
          color: 0x08f70b,
          footer: { text: "Vote below" },
          thumbnail: { url: "{{guild.icon_url}}" },
        },
      ],
    });

    const reactions = find(item, "action_message_reaction_create");
    expect(reactions.map((r) => r.data.emoji_data)).toEqual([
      { name: "👍" },
      { name: "no", id: "123456789012345678" },
    ]);
    expect(reactions[0].data.message_target).toBe(
      `{{result('${message.id}').id}}`
    );

    const [reply] = find(item, "action_response_create");
    expect(reply.data).toEqual({
      message_data: { content: "Sent to {{arg('channel').mention}}!" },
      message_ephemeral: true,
    });
  });

  it("converts conditions with their branches", () => {
    const [item] = convert({
      name: "rank",
      description: "Pick a rank",
      options: [
        {
          type: "string",
          name: "rank",
          description: "Rank",
          required: true,
          choices: ["gold", "silver"],
        },
      ],
      actions: [
        {
          type: "condition",
          condition_type: "option",
          option: "rank",
          conditions: [
            {
              option: "rank",
              value: "gold",
              actions: [
                {
                  type: "role_add",
                  roles: [{ id: "111111111111111111", name: "Gold" }],
                  target: { self: true },
                },
              ],
            },
            {
              option: "rank",
              value: "silver",
              actions: [{ type: "plain_text", response: "Silver it is" }],
            },
          ],
          else: { actions: [{ type: "plain_text", response: "No such rank" }] },
        },
      ],
      permissions: everyone,
    });

    expect(item.status).toBe("ready");
    const [condition] = find(item, "control_condition_compare");
    expect(condition.data.condition_base_value).toBe("{{arg('rank')}}");
    expect(
      find(item, "control_condition_item_compare").map(
        (n) => n.data.condition_item_value
      )
    ).toEqual(["gold", "silver"]);

    const [gold] = find(item, "control_condition_item_compare");
    expect(after(item, gold.id)).toMatchObject({
      type: "action_member_role_add",
      data: { user_target: "{{user.id}}", role_target: "111111111111111111" },
    });
    const [otherwise] = find(item, "control_condition_item_else");
    expect(after(item, otherwise.id)?.data.message_data).toEqual({
      content: "No such rank",
    });
    expect(
      find(item, "option_command_argument")[0].data.command_argument_choices
    ).toEqual([
      { name: "gold", value: "gold" },
      { name: "silver", value: "silver" },
    ]);
  });

  it("converts buttons with the blocks they run", () => {
    const [item] = convert({
      name: "verify",
      description: "Verify yourself",
      options: [],
      actions: [
        {
          type: "button_response",
          response_options: { response: "Press to verify" },
          buttons: [
            {
              label: "Verify",
              style: "SUCCESS",
              actions: [{ type: "plain_text", response: "Done, {user_name}" }],
            },
            {
              label: "Rules",
              style: "LINK",
              URL: "https://example.com/rules",
              actions: [],
            },
          ],
        },
      ],
      permissions: everyone,
    });

    expect(item.status).toBe("ready");
    const [message] = find(item, "action_response_create").filter(
      (n) => (n.data.message_data as { components?: unknown }).components
    );
    expect(message.data.message_data).toEqual({
      content: "Press to verify",
      components: [
        {
          type: 1,
          components: [
            { type: 2, id: 1, style: 3, label: "Verify" },
            {
              type: 2,
              id: 2,
              style: 5,
              label: "Rules",
              url: "https://example.com/rules",
            },
          ],
        },
      ],
    });
    expect(after(item, message.id, "component_1")?.data.message_data).toEqual({
      content: "Done, {{user.username}}",
    });
  });

  it("never aims a punishment at a user it couldn't read", () => {
    const [item] = convert({
      name: "ban",
      description: "Ban someone",
      options: [
        { type: "user", name: "member", description: "Who", required: true },
      ],
      actions: [
        {
          type: "ban_action",
          target: "{option_member}",
          reason: "Banned by {user_name}",
        },
        { type: "kick_action", target: "", reason: "" },
        { type: "kick_action", target: { self: true } },
        { type: "plain_text", response: "Banned {option_member}" },
      ],
      permissions: { ...everyone, required_permissions: ["BAN_MEMBERS"] },
    });

    expect(find(item, "action_member_ban")[0].data).toEqual({
      user_target: "{{arg('member').id}}",
      audit_log_reason: "Banned by {{user.username}}",
    });
    expect(find(item, "action_member_kick")).toEqual([]);
    expect(item.status).toBe("review");
    expect(item.notes.join(" ")).toContain("Kick a Member");
  });

  it("reports blocks and variables Kite has nothing for", () => {
    const [item] = convert({
      name: "coins",
      description: "Your coins",
      options: [],
      actions: [
        { type: "set_var", reference: "coins", value: "5" },
        { type: "plain_text", response: "You have {BGVAR_coins} coins" },
      ],
      permissions: everyone,
    });

    expect(item.status).toBe("review");
    expect(item.notes).toEqual([
      "These BotGhost blocks have no match in Kite yet and were left out: Set Variable.",
      "These BotGhost variables have no placeholder in Kite and were left as text: {BGVAR_coins}.",
    ]);
    const note = item.flow!.nodes.find((n) => n.type === "note");
    expect(note?.data.note_text).toContain("Set Variable");
    expect(blocks(item).map((n) => n.type)).toEqual([
      "entry_command",
      "action_response_create",
    ]);
  });

  it("converts API requests and what reads their response", () => {
    const [item] = convert({
      name: "cat",
      description: "A cat fact",
      options: [],
      actions: [
        {
          type: "api_action",
          name: "fact",
          request_options: {
            method: "GET",
            url: "https://example.com/fact",
            params: [
              { key: "lang", value: "en" },
              { key: "", value: "" },
            ],
            headers: [{ key: "", value: "" }],
            body: [{ key: "", value: "" }],
          },
        },
        { type: "plain_text", response: "{fact.response.data.text}" },
      ],
      permissions: everyone,
    });

    expect(item.status).toBe("ready");
    expect(find(item, "action_http_request")[0].data).toEqual({
      http_request_data: {
        url: "https://example.com/fact?lang=en",
        method: "GET",
      },
      temporary_name: "fact",
    });
    expect(find(item, "action_response_create")[0].data.message_data).toEqual({
      content: "{{var('fact').data().data.text}}",
    });
  });

  it("reads commands that only have the tree of the builder", () => {
    const node = (id: string, type: string, data: unknown) => ({
      id,
      type,
      position: { x: 0, y: 0 },
      data: { node_options: {}, data },
    });
    const edge = (source: string, target: string) => ({
      id: `e-${source}-${target}`,
      source,
      target,
      type: "custom",
    });
    const tree = [
      {
        id: "root",
        type: "root",
        data: { title: "hello", data: {} },
        position: { x: 0, y: 0 },
      },
      node("option_1", "option", {
        type: "user",
        name: "who",
        description: "Who to greet",
        required: true,
      }),
      edge("option_1", "root"),
      node("action_1", "action", {
        type: "plain_text",
        response: "Hi {option_who}",
      }),
      edge("root", "action_1"),
      node("action_2", "action", { type: "wait_action", seconds: 3 }),
      edge("action_1", "action_2"),
      {
        ...node("condition_1", "condition", {}),
        data: { node_options: { condition_type: "chance" }, data: {} },
      },
      edge("action_2", "condition_1"),
      node("child_1", "conditionChild", { value: "25", actions: [] }),
      edge("condition_1", "child_1"),
      node("action_3", "action", { type: "plain_text", response: "Lucky" }),
      edge("child_1", "action_3"),
      node("else_1", "conditionChild", { type: "else", actions: [] }),
      edge("condition_1", "else_1"),
    ];
    const [item] = convert({
      name: "hello",
      description: "Say hi",
      compressed_tree: LZString.compressToBase64(JSON.stringify(tree)),
    });

    expect(item.status).toBe("ready");
    expect(
      blocks(item)
        .map((n) => n.type)
        .sort()
    ).toEqual(
      [
        "entry_command",
        "option_command_argument",
        "action_response_create",
        "control_sleep",
        "action_random_generate",
        "control_condition_compare",
        "control_condition_item_compare",
        "control_condition_item_else",
        "action_response_create",
      ].sort()
    );
    expect(find(item, "action_response_create")[0].data.message_data).toEqual({
      content: "Hi {{arg('who').mention}}",
    });
  });

  it("converts select menus, random replies and role conditions", () => {
    const [item] = convert({
      name: "menu",
      description: "Pick one",
      options: [],
      actions: [
        {
          type: "random_response",
          responses: ["Heads", { response: "Tails" }],
        },
        {
          type: "select_menu",
          response_options: { response: "Pick a color" },
          placeholder: "Colors",
          options: [
            {
              label: "Red",
              description: "Warm",
              actions: [{ type: "plain_text", response: "Red!" }],
            },
            {
              label: "Staff",
              actions: [
                {
                  type: "condition",
                  condition_type: "role",
                  conditions: [
                    {
                      id: "222222222222222222",
                      name: "Staff",
                      actions: [{ type: "plain_text", response: "Hi staff" }],
                    },
                  ],
                  else: {
                    actions: [
                      {
                        type: "timeout_action",
                        target: "{user_id}",
                        duration: "60",
                        reason: "",
                      },
                    ],
                  },
                },
              ],
            },
          ],
        },
      ],
      permissions: everyone,
    });

    // Only the note about the unit of the timeout is left to check.
    expect(item.notes).toEqual([
      "Check the length of the Timeout member block. Kite counts it in seconds.",
    ]);
    expect(find(item, "action_random_generate")[0].data).toMatchObject({
      random_mode: "item",
      random_items: "Heads\nTails",
    });
    const menu = find(item, "action_response_create").find(
      (n) => (n.data.message_data as { components?: unknown }).components
    )!;
    const picked = after(item, menu.id, "component_1")!;
    expect(picked).toMatchObject({
      type: "control_condition_compare",
      data: { condition_base_value: "{{interaction.value}}" },
    });
    expect(
      find(item, "control_condition_compare")[1].data.condition_base_value
    ).toBe('{{join(member.role_ids, ",")}}');
    expect(find(item, "action_member_timeout")[0].data).toEqual({
      user_target: "{{user.id}}",
      member_timeout_duration_seconds: "60",
    });
  });

  it("converts custom events and skips the ones Kite doesn't have", () => {
    const items = convert({
      kite_botghost_export: 1,
      commands: [],
      events: [
        {
          name: "Welcome",
          type: "guildMemberAdd",
          options: [],
          actions: [
            { type: "plain_text", response: "Welcome {user} to {server}!" },
            {
              type: "dm_response",
              response_options: { response: "Read the rules" },
              target: { self: true },
            },
          ],
        },
        {
          name: "Music",
          type: "trackStart",
          options: [],
          actions: [{ type: "plain_text", response: "Playing" }],
        },
      ],
    });

    expect(items.map((i) => [i.kind, i.status])).toEqual([
      ["event", "ready"],
      ["event", "unsupported"],
    ]);
    const [welcome] = items;
    expect(welcome.eventType).toBe("guild_member_add");
    expect(blocks(welcome).map((n) => [n.type, n.data])).toEqual([
      [
        "entry_event",
        {
          event_type: "guild_member_add",
          description: "Imported from BotGhost",
        },
      ],
      [
        "action_message_create",
        {
          channel_target: "{{channel.id}}",
          message_data: {
            content: "Welcome {{user.mention}} to {{guild.name}}!",
          },
        },
      ],
      [
        "action_private_message_create",
        {
          user_target: "{{user.id}}",
          message_data: { content: "Read the rules" },
        },
      ],
    ]);
  });
});

describe("parseBotGhostExport", () => {
  it("finds commands and custom events anywhere in the data", () => {
    const parsed = parseBotGhostExport(
      JSON.stringify({
        bot: {
          commands: {
            slash: {
              commands: [
                { name: "a", description: "A", actions: [], options: [] },
              ],
            },
            custom_events: {
              events: [{ name: "b", type: "messageCreate", actions: [] }],
            },
          },
        },
      })
    );
    expect(parsed.commands.map((c) => c.name)).toEqual(["a"]);
    expect(parsed.events.map((e) => [e.name, e.eventType])).toEqual([
      ["b", "messageCreate"],
    ]);
  });

  it("rejects text that isn't JSON", () => {
    expect(() => parseBotGhostExport("CMD_1234")).toThrow(/BotGhost helper/);
  });
});
