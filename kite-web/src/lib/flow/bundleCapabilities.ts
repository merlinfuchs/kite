import { blockIntegrations, getBlockDefinition } from "../blocks";
import { getIntegration } from "../integrations";
import { placeholderReferences } from "./templateSecrets";
import { isNoteNode } from "./notes";

// What a bundle's blocks can do that someone adding it should know about,
// since its flows run with their bot and credits.
export const capabilityGroups = [
  {
    label: "Bans, unbans, kicks or times out members",
    types: [
      "action_member_ban",
      "action_member_unban",
      "action_member_kick",
      "action_member_timeout",
    ],
  },
  {
    label: "Deletes messages, reactions, channels or roles",
    types: [
      "action_message_delete",
      "action_message_bulk_delete",
      "action_message_reaction_delete",
      "action_channel_delete",
      "action_role_delete",
    ],
  },
  {
    label: "Can make your bot leave a server",
    types: ["action_server_leave"],
  },
  {
    label: "Gives or takes roles, or creates and changes channels and roles",
    types: [
      "action_member_role_add",
      "action_member_role_remove",
      "action_member_edit",
      "action_role_create",
      "action_role_edit",
      "action_channel_create",
      "action_channel_edit",
    ],
  },
  {
    label: "Creates invites to your server",
    types: ["action_invite_create"],
  },
  {
    label: "Sends direct messages",
    types: ["action_private_message_create"],
  },
  {
    label: "Changes your bot's status or server profile",
    types: ["action_status_set", "action_server_profile_edit"],
  },
  {
    label: "Sends requests to other websites",
    types: ["action_http_request"],
  },
  {
    label: "Calls the Discord API directly",
    types: ["action_discord_api_request"],
  },
  {
    label: "Uses AI, which costs more credits",
    types: [
      "action_ai_chat_completion",
      "action_ai_score",
      "action_ai_translate",
      "action_ai_web_search",
      "control_ai_check",
      "control_condition_ai_choice",
    ],
  },
];

// harmlessTypes are the action and control blocks that don't need to be
// pointed out. Every one of them is in a group or here, a test checks it.
export const harmlessTypes = [
  "action_channel_get",
  "action_cookie_api_captcha_create",
  "action_cookie_api_captcha_get",
  "action_cookie_api_card_create",
  "action_cookie_api_minecraft_user_get",
  "action_cookie_api_qr_code_create",
  "action_cookie_api_transcript_create",
  "action_erlc_command_run",
  "action_erlc_server_get",
  "action_expression_evaluate",
  "action_forum_post_create",
  "action_guild_get",
  "action_log",
  "action_member_get",
  "action_message_create",
  "action_message_edit",
  "action_message_forward",
  "action_message_get",
  "action_message_list",
  "action_message_pin",
  "action_message_publish",
  "action_message_reaction_create",
  "action_message_unpin",
  "action_persona_message_create",
  "action_poll_create",
  "action_random_generate",
  "action_response_create",
  "action_response_defer",
  "action_response_delete",
  "action_response_edit",
  "action_roblox_user_get",
  "action_role_get",
  "action_thread_create",
  "action_thread_member_add",
  "action_thread_member_remove",
  "action_typing_trigger",
  "action_user_get",
  "action_value_lookup",
  "action_variable_delete",
  "action_variable_get",
  "action_variable_set",
  "action_voice_channel_join",
  "action_voice_channel_leave",
  "action_webhook_response_create",
  "control_condition_channel",
  "control_condition_compare",
  "control_condition_item_ai_choice",
  "control_condition_item_channel",
  "control_condition_item_compare",
  "control_condition_item_else",
  "control_condition_item_role",
  "control_condition_item_user",
  "control_condition_role",
  "control_condition_user",
  "control_error_handler",
  "control_flow_end",
  "control_if",
  "control_loop",
  "control_loop_each",
  "control_loop_end",
  "control_loop_exit",
  "control_sleep",
  "control_webhook_verify",
];

// bundleCapabilities are the things the blocks of the flows can do, in the
// order of capabilityGroups, then the integrations other than Discord and the
// secrets they use, whether the bundle has IDs of the server it was made in,
// and whether it seems to have a credential.
export function bundleCapabilities(content: SharedContent): string[] {
  const { flows } = content;
  // Notes don't run, what's written in them isn't used.
  const nodes = flows.flatMap((f) => f.nodes).filter((n) => !isNoteNode(n));
  const types = new Set(nodes.map((n) => n.type));

  const res = capabilityGroups.flatMap((g) =>
    g.types.some((t) => types.has(t)) ? [g.label] : []
  );

  const integrations = new Set<string>();
  for (const type of Array.from(types)) {
    const block = getBlockDefinition(type);
    if (!block) continue;
    for (const id of blockIntegrations(block)) {
      if (id !== "discord") integrations.add(id);
    }
  }
  for (const id of Array.from(integrations)) {
    res.push(`Uses ${getIntegration(id)?.name ?? id} with your credentials`);
  }

  // Secrets aren't part of the bundle, but its blocks use the secrets of the
  // same name in the app that adds it.
  const secrets = Array.from(
    new Set(
      blockSettings(flows).flatMap((d) => placeholderReferences(d, "secrets"))
    )
  );
  if (secrets.length) {
    res.push(
      `Reads your ${
        secrets.length === 1 ? "secret" : "secrets"
      } ${new Intl.ListFormat("en").format(secrets)}`
    );
  }

  if (hasServerIds(content)) {
    res.push(
      "Has IDs of channels, roles or users of the server it was made in"
    );
  }
  if (hasCredential(content)) {
    res.push(
      "Seems to have an API key or token typed in, instead of using a secret"
    );
  }
  return res;
}

type Flow = { nodes: { type?: string; data?: unknown }[] };

// SharedContent is what a bundle shares: its flows, the data of its message
// templates and the defaults of its parameters.
type SharedContent = {
  flows: Flow[];
  messages?: unknown[];
  defaults?: string[];
};

// blockSettings are the JSON of the settings of the blocks. Notes don't run,
// so they're left out unless withNotes, but they're part of the bundle.
function blockSettings(flows: Flow[], withNotes = false): string[] {
  return flows
    .flatMap((f) => f.nodes)
    .filter((n) => withNotes || !isNoteNode(n))
    .map((n) => JSON.stringify(n.data ?? {}));
}

// hasServerIds is whether an ID is typed into a block or a message template,
// which only exists in the server it was made in. The IDs in Discord's file
// URLs work everywhere, and the one in a webhook URL is part of a credential.
export function hasServerIds({ flows, messages = [] }: SharedContent): boolean {
  return [
    ...blockSettings(flows),
    ...messages.map((m) => JSON.stringify(m)),
  ].some((d) => /\b\d{17,20}\b/.test(d.replace(urlsWithIds, "")));
}

const urlsWithIds =
  /(cdn|media)\.discordapp\.(com|net)\/[^\s"]*|webhooks\/\d+\/[\w-]*/g;

// hasCredential is whether the bundle seems to have a key or token typed in,
// which anyone can see in it.
export function hasCredential({
  flows,
  messages = [],
  defaults = [],
}: SharedContent): boolean {
  return [
    ...blockSettings(flows, true),
    ...messages.map((m) => JSON.stringify(m)),
    ...defaults,
  ].some((d) => credentialPatterns.some((p) => p.test(d)));
}

// secretName matches the names of headers, query params and JSON fields that
// hold a credential, like X-API-Key or client_secret.
const secretName = String.raw`[\w-]*?(api[-_]?key|token|secret|password|authorization)`;

// credentialPatterns match keys and tokens in the JSON of the settings of
// blocks and message templates, where a secret should be used. A value with a
// placeholder in it isn't one.
const credentialPatterns = [
  // Discord webhook URLs
  /discord(app)?\.com\/api\/(v\d+\/)?webhooks\/\d{17,20}\/[\w-]{60,}/,
  // OpenAI, Anthropic, OpenRouter, Stripe, GitHub, Slack, Google, AWS
  /\b(sk-[\w-]{20,}|[sr]k_live_\w{16,}|gh[pousr]_\w{30,}|github_pat_\w{30,}|xox[abprs]-[\w-]{10,}|AIza[\w-]{35}|AKIA[0-9A-Z]{16})/,
  // Discord bot tokens
  /\b[\w-]{24,28}\.[\w-]{6}\.[\w-]{27,}/,
  /\b(Bearer|Basic) (?=[\w\-.~+/]*\d)[\w\-.~+/]{20,}/,
  // Headers and query params of HTTP requests, fields of JSON bodies and URLs
  new RegExp(`"key":"${secretName}","value":"[^"{]{16,}"`, "i"),
  new RegExp(`"${secretName}":"[^"{]{16,}"`, "i"),
  new RegExp(`[?&#;](key|${secretName})=[^&{"\\s]{16,}`, "i"),
];
