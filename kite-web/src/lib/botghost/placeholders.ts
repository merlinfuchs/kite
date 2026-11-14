// BotGhost writes variables as {user_name}, Kite as {{user.username}}.

export interface PlaceholderContext {
  // The Kite name and type of each command option, by its BotGhost name.
  options: Map<string, { name: string; type: string }>;
  // Names of the API request blocks that ran before, whose response is read
  // as {name.response.field}.
  requests: Set<string>;
  // Variables that have no placeholder in Kite, collected for the report.
  unknown: Set<string>;
}

const simple: Record<string, string> = {
  user: "{{user.mention}}",
  user_id: "{{user.id}}",
  user_name: "{{user.username}}",
  user_username: "{{user.username}}",
  user_tag: "{{user.username}}",
  user_displayName: "{{user.display_name}}",
  user_display_name: "{{user.display_name}}",
  user_nickname: "{{user.display_name}}",
  user_icon: "{{user.avatar_url}}",
  user_avatar: "{{user.avatar_url}}",
  user_createdAt: "<t:{{user.created_at}}:F>",
  user_joined: "<t:{{user.joined_at}}:F>",
  server: "{{guild.name}}",
  server_name: "{{guild.name}}",
  server_id: "{{guild.id}}",
  server_icon: "{{guild.icon_url}}",
  server_members: "{{guild.member_count}}",
  server_member_count: "{{guild.member_count}}",
  server_owner: "<@{{guild.owner_id}}>",
  server_owner_id: "{{guild.owner_id}}",
  channel: "{{channel.mention}}",
  channel_id: "{{channel.id}}",
  channel_name: "{{channel.name}}",
  message_id: "{{message.id}}",
  message_content: "{{message.content}}",
  message_author_id: "{{user.id}}",
  bot_id: "{{app.user.id}}",
  server_count: "{{app.guild_count}}",
  time_now_unix: "{{now().Unix()}}",
};

const variable = /\{([A-Za-z_][A-Za-z0-9_.\-[\],]*)\}/g;

// Options that are a user, channel or role print as a mention in text.
const mentionTypes = ["user", "channel", "role", "mentionable"];

function optionPlaceholder(
  name: string,
  ctx: PlaceholderContext,
  field: "text" | "id"
) {
  const option = ctx.options.get(name);
  if (!option) return null;
  if (mentionTypes.includes(option.type)) {
    return `{{arg('${option.name}').${field === "id" ? "id" : "mention"}}}`;
  }
  return `{{arg('${option.name}')}}`;
}

function replaceOne(
  name: string,
  ctx: PlaceholderContext,
  field: "text" | "id"
): string | null {
  if (name.startsWith("option_")) {
    return optionPlaceholder(name.slice("option_".length), ctx, field);
  }
  if (field === "id" && name === "user") return "{{user.id}}";
  if (field === "id" && name === "channel") return "{{channel.id}}";
  if (Object.hasOwn(simple, name)) return simple[name];

  const [request, part, ...path] = name.split(".");
  if (ctx.requests.has(request) && part === "response") {
    const safe = path.every((p) => /^[A-Za-z_][A-Za-z0-9_]*$/.test(p));
    if (safe) {
      return `{{var('${request}').data()${path.map((p) => `.${p}`).join("")}}}`;
    }
  }
  if (ctx.requests.has(request) && part === "status" && path.length === 0) {
    return `{{var('${request}').status_code}}`;
  }
  return null;
}

// translateText replaces the BotGhost variables of a text with Kite
// placeholders. Variables Kite has nothing for stay as they are, so they can
// be found in the editor.
export function translateText(text: unknown, ctx: PlaceholderContext): string {
  if (typeof text !== "string") return "";
  return text.replace(variable, (match, name: string) => {
    const replaced = replaceOne(name, ctx, "text");
    if (replaced === null) ctx.unknown.add(match);
    return replaced ?? match;
  });
}

// translateId reads a setting that holds an ID: digits, or one variable. It's
// null if it's neither, as Kite's ID settings take nothing else.
export function translateId(
  value: unknown,
  ctx: PlaceholderContext
): string | null {
  if (typeof value !== "string") return null;
  const text = value.trim();
  if (/^[0-9]{5,25}$/.test(text)) return text;

  const match = /^\{([^{}]+)\}$/.exec(text);
  // Targets are sometimes stored as the bare name of an option.
  const name = match ? match[1] : text;
  if (!match && !ctx.options.has(name)) return null;

  const replaced = match
    ? replaceOne(name, ctx, "id")
    : optionPlaceholder(name, ctx, "id");
  if (replaced === null || !/^\{\{[^{}]+\}\}$/.test(replaced)) {
    if (match) ctx.unknown.add(text);
    return null;
  }
  return replaced;
}
