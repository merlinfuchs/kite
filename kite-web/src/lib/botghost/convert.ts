import { Node } from "@xyflow/react";
import { NodeData } from "../flow/dataSchema";
import { FlowContextType } from "../flow/context";
import { applyFlowEdits, FlowEdit } from "../flow/edits";
import { newCommandFlow, newEventListenerFlow } from "../flow/newFlows";
import { getNodeId } from "../flow/nodes";
import { defaultNoteSize, noteMaxLength, noteNodeType } from "../flow/notes";
import {
  encodePermissionsBitset,
  permissionBits,
} from "../discord/permissions";
import { CommandType } from "../flow/commandTypes";
import { FlowData } from "../types/flow.gen";
import { BotGhostAction, BotGhostItem, isRecord } from "./parse";
import { PlaceholderContext, translateId, translateText } from "./placeholders";

// What became of a BotGhost command or custom event:
// - ready: every block has a Kite block that does the same.
// - review: it was converted, but something needs a look in the editor first.
// - unsupported: there's nothing to import, like an event Kite doesn't have.
export type ConvertStatus = "ready" | "review" | "unsupported";

export interface ConvertedItem {
  kind: "command" | "event";
  // The name in BotGhost, and the one it gets in Kite.
  sourceName: string;
  name: string;
  description: string;
  // The Kite event type of an event listener.
  eventType: string;
  status: ConvertStatus;
  // What the user should check, also put on the canvas as a note.
  notes: string[];
  flow: FlowData | null;
}

// The Kite event of each BotGhost custom event that has one.
const eventTypes: Record<string, string> = {
  guildMemberAdd: "guild_member_add",
  guildMemberRemove: "guild_member_remove",
  messageCreate: "message_create",
  messageUpdate: "message_update",
  messageDelete: "message_delete",
  messageReactionAdd: "message_reaction_add",
  messageReactionRemove: "message_reaction_remove",
  serverJoin: "guild_create",
  guildDelete: "guild_delete",
  guildMemberBoost: "guild_boost",
  guildMemberUnBoost: "guild_boost_remove",
  voiceChannelJoin: "voice_state_update",
  voiceChannelSwitch: "voice_state_update",
  voiceChannelLeave: "voice_state_update",
};

const eventHints: Record<string, string> = {
  timedEvent: "Recreate it under Scheduled tasks.",
  webhookExecuted: "Recreate it under Webhooks.",
};

const argumentTypes: Record<string, string> = {
  string: "string",
  int: "integer",
  integer: "integer",
  number: "number",
  user: "user",
  channel: "channel",
  role: "role",
  boolean: "boolean",
  attachment: "attachment",
};

const buttonStyles: Record<string, number> = {
  PRIMARY: 1,
  SECONDARY: 2,
  SUCCESS: 3,
  DANGER: 4,
  LINK: 5,
};

const compareModes: Record<string, string> = {
  equal_to: "equal",
  not: "not_equal",
  greater_than: "greater_than",
  greater_than_equal_to: "greater_than_or_equal",
  less_than: "less_than",
  less_than_equal_to: "less_than_or_equal",
  contains: "contains",
  starts_with: "starts_with",
  ends_with: "ends_with",
};

const intervalSeconds: Record<string, number> = {
  seconds: 1,
  minutes: 60,
  hours: 3600,
  days: 86400,
};

// Names for the blocks that aren't imported, as BotGhost shows them.
const blockNames: Record<string, string> = {
  set_var: "Set Variable",
  math_var: "Run Equation on Variable",
  set_unique_variable: "Set Unique Variable",
  delete_message: "Delete a Message",
  react_message: "React to a Message",
  pin_message: "Pin a Message",
  purge_messages: "Purge Messages",
  modal: "Send a Form",
  loop: "Run a Loop",
  channel_create: "Create a Channel",
  channel_edit: "Edit a Channel",
  channel_delete: "Delete a Channel",
  create_role: "Create a Role",
  edit_role: "Edit a Role",
  delete_role: "Delete a Role",
  thread_create: "Create a Thread",
  status_change: "Change the Bot's Status",
  trigger_custom_event: "Trigger a Custom Event",
  string_manipulation: "Manipulate some text",
  ifttt_action: "Trigger an IFTTT event",
  create_invite: "Create an Invite",
  send_error_message: "Send an Error Message",
  welcome_card: "Make an Image Card",
};

// Groups of blocks that share a start, like music_play and music_pause.
const blockGroups: Record<string, string> = {
  music_: "Music",
  ticket_: "Tickets",
  voice_: "Voice",
  scheduled_event_: "Scheduled events",
  forum_post_: "Forum posts",
  thread_: "Threads",
};

function blockName(type: string) {
  const group = Object.keys(blockGroups).find((g) => type.startsWith(g));
  return (
    blockNames[type] ??
    (group ? blockGroups[group] : type.replace(/_/g, " ") || "Unknown")
  );
}

interface Cursor {
  after: string;
  handle?: string;
}

function str(value: unknown) {
  return typeof value === "string" ? value : "";
}

function list(value: unknown): Record<string, unknown>[] {
  return Array.isArray(value) ? value.filter(isRecord) : [];
}

function rec(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {};
}

class Converter {
  edits: FlowEdit[] = [];
  notes: string[] = [];
  skipped = new Set<string>();
  ctx: PlaceholderContext = {
    options: new Map(),
    requests: new Set(),
    unknown: new Set(),
  };
  private count = 0;
  // Above 0 while the blocks of a button or select menu are converted, which
  // answer the click instead of the command or event.
  private componentDepth = 0;

  constructor(private kind: "command" | "event", private ephemeral: boolean) {}

  note(text: string) {
    if (!this.notes.includes(text)) this.notes.push(text);
  }

  private text(value: unknown, max?: number) {
    const text = translateText(value, this.ctx);
    return max ? text.slice(0, max) : text;
  }

  add(type: string, data: NodeData, cursor: Cursor, items?: NodeData[]) {
    const ref = `$b${++this.count}`;
    this.edits.push({
      op: "add_node",
      ref,
      type,
      data,
      after: cursor.after,
      handle: cursor.handle,
      ...(items ? { items } : {}),
    });
    return ref;
  }

  option(type: string, data: NodeData) {
    this.edits.push({ op: "add_node", ref: `$b${++this.count}`, type, data });
  }

  actions(actions: unknown, start: Cursor): Cursor {
    let cursor = start;
    const items = list(actions);
    for (const [i, action] of items.entries()) {
      const next = this.action(action, cursor);
      // Nothing runs after a condition in Kite, only after its branches.
      if (next === null) {
        if (i < items.length - 1) {
          this.note(
            "Blocks that came after a condition weren't imported. Add them to the branches they belong to."
          );
        }
        break;
      }
      cursor = next;
    }
    return cursor;
  }

  // Returns where the next block goes, or null after a condition.
  private action(a: BotGhostAction, cursor: Cursor): Cursor | null {
    const type = a.type === "comparison" ? "condition" : str(a.type);
    switch (type) {
      case "plain_text":
        return this.send(
          { content: a.response, emojis: a.emojis },
          a.target,
          cursor
        );
      case "embed":
        return this.send(
          { embeds: [a.embed], emojis: a.emojis },
          a.target,
          cursor
        );
      case "targeted_response": {
        const o = rec(a.response_options);
        return this.send(
          { content: o.response, embeds: [o.embed], emojis: o.emojis },
          a.target ?? o.target,
          cursor
        );
      }
      case "advanced_message": {
        const m = rec(a.message);
        if (a.useComponentsV2 === true) {
          this.note(
            "A message that uses Components V2 was imported without its layout. Rebuild it in the message editor."
          );
        }
        return this.send(
          {
            content: m.content ?? m.response,
            embeds: Array.isArray(m.embeds) ? m.embeds : [m.embed],
            buttons: [...list(a.buttons), ...list(a.components)],
          },
          a.target,
          cursor
        );
      }
      case "button_response": {
        const o = rec(a.response_options);
        return this.send(
          {
            content: o.response,
            embeds: [o.embed],
            emojis: o.emojis,
            buttons: list(a.buttons),
          },
          a.target ?? o.target,
          cursor
        );
      }
      case "select_menu": {
        const o = rec(a.response_options);
        return this.send(
          {
            content: o.response,
            embeds: [o.embed],
            select: { placeholder: a.placeholder, options: list(a.options) },
          },
          a.target ?? o.target,
          cursor
        );
      }
      case "dm_response": {
        const o = rec(a.response_options);
        const user = this.user(a.target, true);
        const data = this.message({ content: o.response, embeds: [o.embed] });
        if (!user || !data) return this.skip("Send a Direct Message", cursor);
        return {
          after: this.add(
            "action_private_message_create",
            { user_target: user, message_data: data },
            cursor
          ),
        };
      }
      case "random_response":
        return this.random(a, cursor);
      case "role_add":
      case "role_remove":
        return this.roles(a, type, cursor);
      case "kick_action":
      case "ban_action":
      case "unban_action": {
        const user = this.user(a.target, false);
        const title = {
          kick_action: "Kick a Member",
          ban_action: "Ban a Member",
          unban_action: "Unban a Member",
        }[type];
        if (!user) return this.skip(title, cursor);
        const reason = this.text(a.reason, 512);
        return {
          after: this.add(
            {
              kick_action: "action_member_kick",
              ban_action: "action_member_ban",
              unban_action: "action_member_unban",
            }[type],
            {
              user_target: user,
              ...(reason ? { audit_log_reason: reason } : {}),
            },
            cursor
          ),
        };
      }
      case "timeout_action": {
        const user = this.user(a.target, false);
        const duration = this.number(a.duration);
        if (!user || !duration || a.remove === true) {
          return this.skip("Timeout a Member", cursor);
        }
        const reason = this.text(a.reason, 512);
        this.note(
          "Check the length of the Timeout member block. Kite counts it in seconds."
        );
        return {
          after: this.add(
            "action_member_timeout",
            {
              user_target: user,
              member_timeout_duration_seconds: duration,
              ...(reason ? { audit_log_reason: reason } : {}),
            },
            cursor
          ),
        };
      }
      case "change_nickname": {
        const user = this.user(a.target, true);
        const nick = this.text(a.name, 32);
        if (!user || !nick) return this.skip("Change a Nickname", cursor);
        return {
          after: this.add(
            "action_member_edit",
            { user_target: user, member_data: { nick } },
            cursor
          ),
        };
      }
      case "wait_action": {
        const seconds = this.number(a.seconds);
        if (!seconds)
          return this.skip("Wait before running another action", cursor);
        if (Number(seconds) > 25) {
          this.note(
            "A Wait block is longer than 25 seconds. A flow in Kite stops after 30 seconds, so shorten it."
          );
        }
        return {
          after: this.add(
            "control_sleep",
            { sleep_duration_seconds: seconds },
            cursor
          ),
        };
      }
      case "api_action":
        return this.request(a, cursor);
      case "condition":
        return this.condition(a, cursor);
      case "note_action":
        return cursor;
      default:
        return this.skip(blockName(type), cursor);
    }
  }

  private skip(title: string, cursor: Cursor) {
    this.skipped.add(title);
    return cursor;
  }

  // A number setting: digits, or a single placeholder.
  private number(value: unknown) {
    if (typeof value === "number" && value >= 0)
      return String(Math.round(value));
    const text = str(value).trim();
    if (/^[0-9]+$/.test(text)) return text;
    const id = /^\{[^{}]+\}$/.test(text) ? this.text(text) : "";
    return /^\{\{[^{}]+\}\}$/.test(id) ? id : null;
  }

  // The user a block acts on. Blocks that punish never fall back to the user
  // who ran the command, so a target that can't be read skips the block.
  private user(target: unknown, allowSelf: boolean) {
    if (isRecord(target)) {
      if (target.self === true) return allowSelf ? "{{user.id}}" : null;
      return (
        translateId(target.id, this.ctx) ??
        translateId(target.user_id, this.ctx) ??
        translateId(target.variable, this.ctx)
      );
    }
    if (target === undefined || target === "trigger") {
      return allowSelf ? "{{user.id}}" : null;
    }
    return translateId(target, this.ctx);
  }

  // Where a message goes: as the reply, or to a channel.
  private channel(target: unknown): "reply" | string {
    const here =
      this.kind === "event" && this.componentDepth === 0
        ? "{{channel.id}}"
        : "reply";
    if (!isRecord(target)) return here;
    if (target.trigger_channel === true) return "{{channel.id}}";
    if ("editId" in target || target.action === true) {
      this.note(
        "A block that edits a message sent earlier was imported as a new message."
      );
      return here;
    }
    return (
      translateId(target.id, this.ctx) ??
      translateId(target.variable, this.ctx) ??
      here
    );
  }

  private embed(value: unknown) {
    const e = rec(value);
    const out: Record<string, unknown> = {};
    const url = (v: unknown) => this.text(isRecord(v) ? v.url : v);

    if (e.title) out.title = this.text(e.title, 256);
    if (e.description) out.description = this.text(e.description, 4096);
    if (typeof e.url === "string" && e.url) out.url = this.text(e.url);
    if (typeof e.color === "number") out.color = e.color;
    if (typeof e.color === "string" && /^#?[0-9a-f]{6}$/i.test(e.color)) {
      out.color = parseInt(e.color.replace("#", ""), 16);
    }
    const footer = isRecord(e.footer) ? e.footer : { text: e.footer };
    if (footer.text) {
      const icon = this.text(footer.icon_url ?? footer.iconURL);
      out.footer = {
        text: this.text(footer.text, 2048),
        ...(icon ? { icon_url: icon } : {}),
      };
    }
    const author = isRecord(e.author) ? e.author : { name: e.author };
    if (author.name) {
      const icon = this.text(author.icon_url ?? author.iconURL);
      out.author = {
        name: this.text(author.name, 256),
        ...(icon ? { icon_url: icon } : {}),
      };
    }
    if (url(e.thumbnail)) out.thumbnail = { url: url(e.thumbnail) };
    if (url(e.image)) out.image = { url: url(e.image) };
    const fields = list(e.fields)
      .filter((f) => f.name && f.value)
      .slice(0, 25)
      .map((f) => ({
        name: this.text(f.name, 256),
        value: this.text(f.value, 1024),
        ...(f.inline ? { inline: true } : {}),
      }));
    if (fields.length > 0) out.fields = fields;

    return Object.keys(out).length > 0 ? out : null;
  }

  private message(msg: { content?: unknown; embeds?: unknown[] }) {
    const content = this.text(msg.content, 2000);
    const embeds = (msg.embeds ?? [])
      .flatMap((e) => this.embed(e) ?? [])
      .slice(0, 10);
    if (!content && embeds.length === 0) return null;
    return {
      ...(content ? { content } : {}),
      ...(embeds.length > 0 ? { embeds } : {}),
    } as Record<string, unknown>;
  }

  private send(
    msg: {
      content?: unknown;
      embeds?: unknown[];
      emojis?: unknown;
      buttons?: Record<string, unknown>[];
      select?: { placeholder: unknown; options: Record<string, unknown>[] };
    },
    target: unknown,
    cursor: Cursor
  ): Cursor {
    const data = this.message(msg);
    if (!data) {
      this.note("A message without text or an embed wasn't imported.");
      return cursor;
    }

    // Buttons and select menus, with the blocks that run when they're used.
    const branches: { id: number; run: (cursor: Cursor) => void }[] = [];
    const rows: Record<string, unknown>[] = [];
    let id = 0;

    const buttons = (msg.buttons ?? []).flatMap((b) => {
      const style =
        buttonStyles[str(b.style).toUpperCase()] ?? buttonStyles.PRIMARY;
      const label = this.text(b.label, 80) || "Button";
      if (b.disableOnUse || b.cooldownEnabled) {
        this.note(
          "Buttons that turn off after a click or have a cooldown were imported as normal buttons."
        );
      }
      if (style === buttonStyles.LINK) {
        const url = this.text(b.URL ?? b.url);
        return url ? [{ type: 2, id: ++id, style, label, url }] : [];
      }
      const button = { type: 2, id: ++id, style, label };
      branches.push({
        id,
        run: (c) => void this.actions(b.actions, c),
      });
      return [button];
    });
    for (let i = 0; i < buttons.length && rows.length < 5; i += 5) {
      rows.push({ type: 1, components: buttons.slice(i, i + 5) });
    }

    const options = (msg.select?.options ?? []).slice(0, 25);
    if (options.length > 0 && rows.length < 5) {
      const selectId = ++id;
      const placeholder = this.text(msg.select?.placeholder, 150);
      rows.push({
        type: 1,
        components: [
          {
            type: 3,
            id: selectId,
            ...(placeholder ? { placeholder } : {}),
            min_values: 1,
            max_values: 1,
            options: options.map((o, i) => {
              const description = this.text(o.description, 100);
              return {
                label: this.text(o.label, 100) || `Option ${i + 1}`,
                value: `option_${i + 1}`,
                ...(description ? { description } : {}),
              };
            }),
          },
        ],
      });
      branches.push({
        id: selectId,
        run: (c) => {
          const ref = this.add(
            "control_condition_compare",
            { condition_base_value: "{{interaction.value}}" },
            c,
            options.map((_, i) => ({
              condition_item_mode: "equal",
              condition_item_value: `option_${i + 1}`,
            }))
          );
          options.forEach((o, i) =>
            this.actions(o.actions, { after: `${ref}.item${i}` })
          );
        },
      });
    }
    if (rows.length > 0) data.components = rows;

    const where = this.channel(target);
    const ref =
      where === "reply"
        ? this.add(
            "action_response_create",
            {
              message_data: data,
              ...(this.ephemeral ? { message_ephemeral: true } : {}),
            },
            cursor
          )
        : this.add(
            "action_message_create",
            { channel_target: where, message_data: data },
            cursor
          );

    let last = ref;
    for (const emoji of (Array.isArray(msg.emojis) ? msg.emojis : []).slice(
      0,
      10
    )) {
      const custom = /^<a?:(\w+):(\d+)>$/.exec(str(emoji).trim());
      if (!custom && !str(emoji).trim()) continue;
      last = this.add(
        "action_message_reaction_create",
        {
          channel_target: `{{result('${ref}').channel_id}}`,
          message_target: `{{result('${ref}').id}}`,
          emoji_data: custom
            ? { name: custom[1], id: custom[2] }
            : { name: str(emoji).trim() },
        },
        { after: last }
      );
    }

    this.componentDepth++;
    for (const branch of branches) {
      branch.run({ after: ref, handle: `component_${branch.id}` });
    }
    this.componentDepth--;

    return { after: last };
  }

  private random(a: BotGhostAction, cursor: Cursor): Cursor {
    const replies = (Array.isArray(a.responses) ? a.responses : [])
      .map((r) => (isRecord(r) ? r.response : r))
      .map((r) =>
        this.text(r)
          .replace(/\s*\n\s*/g, " ")
          .trim()
      )
      .filter(Boolean);
    if (replies.length === 0) return this.skip("Send a Random Reply", cursor);

    const name = `random_reply_${++this.count}`;
    const pick = this.add(
      "action_random_generate",
      {
        random_mode: "item",
        random_items: replies.join("\n").slice(0, 10000),
        temporary_name: name,
      },
      cursor
    );
    return this.send({ content: `{{var('${name}')}}` }, a.target, {
      after: pick,
    });
  }

  private roles(a: BotGhostAction, type: string, cursor: Cursor): Cursor {
    const title = type === "role_add" ? "Add Roles" : "Remove Roles";
    const user = this.user(a.target, true);
    const roles = [
      ...(Array.isArray(a.roles) ? a.roles : []).map((r) =>
        isRecord(r) ? r.id ?? r.variable : r
      ),
      ...str(a.additional_roles).split(/[\s,]+/),
    ].flatMap((r) => translateId(r, this.ctx) ?? []);
    if (!user || roles.length === 0) return this.skip(title, cursor);

    let after = cursor;
    for (const role of [...new Set(roles)]) {
      after = {
        after: this.add(
          type === "role_add"
            ? "action_member_role_add"
            : "action_member_role_remove",
          { user_target: user, role_target: role },
          after
        ),
      };
    }
    return after;
  }

  private request(a: BotGhostAction, cursor: Cursor): Cursor {
    const options = rec(a.request_options);
    const method = str(options.method).toUpperCase() || "GET";
    const pairs = (value: unknown) =>
      list(value)
        .filter((p) => str(p.key).trim())
        .map((p) => ({ key: str(p.key).trim(), value: this.text(p.value) }));

    let url = this.text(options.url).trim();
    const params = pairs(options.params);
    if (params.length > 0) {
      url +=
        (url.includes("?") ? "&" : "?") +
        params.map((p) => `${p.key}=${p.value}`).join("&");
    }
    if (
      !/^https?:\/\//.test(url) ||
      !["GET", "POST", "PUT", "PATCH", "DELETE"].includes(method)
    ) {
      return this.skip("Send an API Request", cursor);
    }

    const headers = pairs(options.headers);
    if (
      headers.some((h) => /TOKEN_SECRET|authorization/i.test(h.key + h.value))
    ) {
      this.note(
        "An API request sends a key or token. Store it under Secrets and check the request's headers."
      );
    }
    const body = Object.fromEntries(
      pairs(options.body).map((p) => [p.key, p.value])
    );

    const name = str(a.name).trim();
    const temporary = /^[a-z0-9_]{1,32}$/.test(name) ? name : "";
    if (temporary) this.ctx.requests.add(name);
    else if (name) {
      this.note(
        `The response of the API request "${name}" is used by later blocks. Give the block a temporary name and update them.`
      );
    }

    return {
      after: this.add(
        "action_http_request",
        {
          http_request_data: {
            url,
            method,
            ...(headers.length > 0 ? { headers } : {}),
            ...(method !== "GET" && Object.keys(body).length > 0
              ? { body_json: body }
              : {}),
          },
          ...(temporary ? { temporary_name: temporary } : {}),
        },
        cursor
      ),
    };
  }

  private condition(a: BotGhostAction, cursor: Cursor): Cursor | null {
    const kind = str(a.condition_type);
    const branches = list(a.conditions);
    const otherwise = rec(a.else).actions;
    const value = (b: Record<string, unknown>) =>
      isRecord(b.value) ? b.value.id : b.value ?? b.id;

    let type = "control_condition_compare";
    let base = "";
    let after = cursor;
    let items: NodeData[] | null = null;

    if (kind === "option") {
      const name = str(branches[0]?.option ?? a.option);
      const option = this.ctx.options.get(name);
      if (option) {
        const byId = ["user", "channel", "role"].includes(option.type);
        base = `{{arg('${option.name}')${byId ? ".id" : ""}}}`;
        items = branches.map((b) => ({
          condition_item_mode: "equal",
          condition_item_value: this.text(String(value(b) ?? "")),
        }));
      }
    } else if (kind === "comparison") {
      base = this.text(a.base_value);
      items = branches.map((b) => {
        const mode = compareModes[str(b.type ?? b.comparison_type)];
        if (!mode) {
          this.note(
            "A Comparison Condition has a branch Kite compares differently. Check how its branches compare."
          );
        }
        return {
          condition_item_mode: mode ?? "equal",
          condition_item_value: this.text(
            String(b.comparison_value ?? b.value ?? "")
          ),
        };
      });
    } else if (kind === "channel" || kind === "user") {
      type = `control_condition_${kind}`;
      base = kind === "channel" ? "{{channel.id}}" : "{{user.id}}";
      const ids = branches.map((b) =>
        translateId(String(value(b) ?? ""), this.ctx)
      );
      if (ids.every((id) => id !== null)) {
        items = ids.map((id) => ({
          condition_item_mode: "equal",
          condition_item_value: id,
        }));
      }
    } else if (kind === "role" || kind === "permissions") {
      // Kite compares one value per condition, so the roles or permissions
      // of the user are joined and each branch looks for its own in them.
      base =
        kind === "role"
          ? '{{join(member.role_ids, ",")}}'
          : '{{join(user.permissions, ",")}}';
      const values = branches.map((b) =>
        kind === "role"
          ? translateId(String(value(b) ?? ""), this.ctx)
          : str(b.permission ?? value(b)).toLowerCase() || null
      );
      if (values.every((v) => v !== null)) {
        items = values.map((v) => ({
          condition_item_mode: "contains",
          condition_item_value: v,
        }));
      }
    } else if (kind === "chance") {
      const name = `chance_${++this.count}`;
      after = {
        after: this.add(
          "action_random_generate",
          { random_min: "1", random_max: "101", temporary_name: name },
          cursor
        ),
      };
      base = `{{var('${name}')}}`;
      let total = 0;
      items = branches.map((b) => {
        total += Number(value(b)) || 0;
        return {
          condition_item_mode: "less_than_or_equal",
          condition_item_value: String(Math.min(total, 100)),
        };
      });
    }

    if (items === null) {
      const title = kind ? `${kind.replace(/_/g, " ")} condition` : "condition";
      this.note(
        `A ${title} wasn't imported, and neither were the blocks of its branches. Rebuild it with a Comparison Condition.`
      );
      return null;
    }
    // Without branches only what's under Else runs.
    if (items.length === 0) {
      this.actions(otherwise, after);
      return null;
    }

    const ref = this.add(type, { condition_base_value: base }, after, items);
    branches.forEach((b, i) =>
      this.actions(b.actions, { after: `${ref}.item${i}` })
    );
    this.actions(otherwise, { after: `${ref}.else` });
    return null;
  }
}

function commandName(name: string, type: CommandType) {
  if (type !== "chat_input") return name.slice(0, 32);
  return (
    name
      .toLowerCase()
      .replace(/^\//, "")
      .split(/\s+/)
      .map((word) => word.replace(/[^\p{L}\p{N}_-]/gu, "").slice(0, 32))
      .filter(Boolean)
      .slice(0, 3)
      .join(" ") || "imported"
  );
}

function argumentName(name: string, taken: Set<string>) {
  const base =
    name
      .toLowerCase()
      .replace(/[^a-z0-9_]/g, "_")
      .slice(0, 28) || "option";
  let unique = base;
  for (let i = 2; taken.has(unique); i++) unique = `${base}_${i}`;
  taken.add(unique);
  return unique;
}

// convertBotGhostItem turns a BotGhost command or custom event into a Kite
// flow, as close to what it did as Kite's blocks allow.
export function convertBotGhostItem(item: BotGhostItem): ConvertedItem {
  const { raw } = item;
  const base = {
    kind: item.kind,
    sourceName: item.name,
    name: item.name,
    description: item.description.slice(0, 100) || "Imported from BotGhost",
    eventType: "",
  };

  if (item.kind === "event" && !eventTypes[item.eventType]) {
    return {
      ...base,
      status: "unsupported",
      notes: [
        `Kite has no event listener for this event yet. ${
          eventHints[item.eventType] ?? ""
        }`.trim(),
      ],
      flow: null,
    };
  }

  const converter = new Converter(item.kind, raw.ephemeral === true);
  const commandType: CommandType =
    raw.command_type === "user" || raw.command_type === "message"
      ? raw.command_type
      : "chat_input";
  const name =
    item.kind === "command" ? commandName(item.name, commandType) : item.name;
  const eventType = eventTypes[item.eventType] ?? "";

  if (item.kind === "command" && commandType === "chat_input") {
    const taken = new Set<string>();
    for (const option of item.options) {
      const source = str(option.name);
      const type = argumentTypes[str(option.type)];
      if (!source || !type) continue;
      const argument = argumentName(source, taken);
      converter.ctx.options.set(source, { name: argument, type });

      const choices = (Array.isArray(option.choices) ? option.choices : [])
        .map((c) => (isRecord(c) ? c : { name: c, value: c }))
        .map((c) => ({
          name: String(c.name ?? "").slice(0, 100),
          value: String(c.value ?? c.name ?? "").slice(0, 100),
        }))
        .filter((c) => c.name && c.value)
        .slice(0, 25);
      converter.option("option_command_argument", {
        name: argument,
        description: str(option.description).trim().slice(0, 100) || argument,
        command_argument_type: type,
        ...(option.required ? { command_argument_required: true } : {}),
        ...(choices.length > 0 && ["string", "integer"].includes(type)
          ? { command_argument_choices: choices }
          : {}),
      });
    }

    const permissions = rec(raw.permissions);
    const bits = (
      Array.isArray(permissions.required_permissions)
        ? permissions.required_permissions
        : []
    ).flatMap((flag) => permissionBits.find((p) => p.flag === flag)?.bit ?? []);
    if (bits.length > 0) {
      converter.option("option_command_permissions", {
        command_permissions: encodePermissionsBitset(bits),
      });
    }
    const restricted =
      list(permissions.allowed_roles).some((r) => r.id !== "everyone") ||
      ["banned_roles", "banned_channels", "banned_users"].some(
        (key) => Array.isArray(permissions[key]) && permissions[key].length > 0
      );
    if (restricted) {
      converter.note(
        "The roles, channels and users that could use this command in BotGhost weren't imported. Set them in Discord under Server Settings, Integrations."
      );
    }

    const cooldown = rec(raw.cooldown);
    if (cooldown.type && cooldown.type !== "disabled") {
      const seconds =
        Number(cooldown.interval) *
        (intervalSeconds[str(cooldown.interval_type)] ?? 0);
      if (seconds >= 1 && seconds <= 3600) {
        converter.option("option_command_cooldown", {
          cooldown_scope:
            cooldown.type === "server" || cooldown.type === "guild"
              ? "guild"
              : cooldown.type === "global"
              ? "global"
              : "user",
          cooldown_duration_seconds: String(Math.round(seconds)),
        });
      } else {
        converter.note(
          "The cooldown wasn't imported. Cooldowns in Kite last up to 1 hour."
        );
      }
    }
  }

  if (eventType === "voice_state_update") {
    converter.note(
      "This event listener runs when someone joins, leaves or moves between voice channels. Add an Event Filter to pick one."
    );
  }

  const blocksBefore = converter.edits.length;
  converter.actions(item.actions, { after: "$entry" });
  const blocks = converter.edits.length - blocksBefore;

  if (blocks === 0) {
    return {
      ...base,
      name,
      eventType,
      status: "unsupported",
      notes: [
        item.actions.length === 0
          ? "It has no blocks."
          : `None of its blocks have a match in Kite yet: ${[
              ...converter.skipped,
            ].join(", ")}.`,
      ],
      flow: null,
    };
  }

  if (converter.skipped.size > 0) {
    converter.note(
      `These BotGhost blocks have no match in Kite yet and were left out: ${[
        ...converter.skipped,
      ].join(", ")}.`
    );
  }
  if (converter.ctx.unknown.size > 0) {
    converter.note(
      `These BotGhost variables have no placeholder in Kite and were left as text: ${[
        ...converter.ctx.unknown,
      ].join(", ")}.`
    );
  }

  const context: FlowContextType =
    item.kind === "command" ? "command" : "event_discord";
  const start =
    item.kind === "command"
      ? newCommandFlow(name, base.description, commandType)
      : newEventListenerFlow(eventType, base.description);
  const result = applyFlowEdits(
    { nodes: start.nodes as Node<NodeData>[], edges: [] },
    converter.edits,
    context
  );
  for (const issue of result.issues) {
    if (issue.severity === "error") {
      converter.note(
        `Kite found a problem to fix in the editor: ${issue.message}`
      );
    }
  }

  const nodes = result.nodes.map((n) => ({
    id: n.id,
    type: n.type,
    position: n.position,
    data: n.data,
  }));
  if (converter.notes.length > 0) {
    nodes.push(importNote(converter.notes, result.nodes));
  }

  return {
    ...base,
    name,
    eventType,
    status: converter.notes.length > 0 ? "review" : "ready",
    notes: converter.notes,
    flow: {
      nodes,
      edges: result.edges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        ...(e.sourceHandle ? { sourceHandle: e.sourceHandle } : {}),
        ...(e.targetHandle ? { targetHandle: e.targetHandle } : {}),
        ...(e.type ? { type: e.type } : {}),
      })),
    } as FlowData,
  };
}

// The note on the canvas that lists what to check, right of the blocks.
function importNote(notes: string[], nodes: Node<NodeData>[]) {
  const right = Math.max(...nodes.map((n) => n.position.x));
  const top = Math.min(...nodes.map((n) => n.position.y));
  return {
    id: getNodeId(noteNodeType),
    type: noteNodeType,
    position: { x: right + 500, y: top },
    width: defaultNoteSize.width + 140,
    height: defaultNoteSize.height + 140,
    data: {
      note_title: "Imported from BotGhost",
      note_text: notes
        .map((n) => `- ${n}`)
        .join("\n")
        .slice(0, noteMaxLength),
    },
  };
}

export function convertBotGhostItems(items: BotGhostItem[]) {
  return items.map(convertBotGhostItem);
}
