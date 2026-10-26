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
    label: "Changes your bot's status",
    types: ["action_status_set"],
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
  "control_loop",
  "control_loop_each",
  "control_loop_end",
  "control_loop_exit",
  "control_sleep",
  "control_webhook_verify",
];

// bundleCapabilities are the things the blocks of the flows can do, in the
// order of capabilityGroups, then the integrations other than Discord and the
// secrets they use, and whether they have IDs of the server they were made in.
export function bundleCapabilities(
  flows: { nodes: { type?: string; data?: unknown }[] }[]
): string[] {
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
  const data = nodes.map((n) => JSON.stringify(n.data ?? {}));
  const secrets = Array.from(
    new Set(data.flatMap((d) => placeholderReferences(d, "secrets")))
  );
  if (secrets.length) {
    res.push(
      `Reads your ${
        secrets.length === 1 ? "secret" : "secrets"
      } ${new Intl.ListFormat("en").format(secrets)}`
    );
  }

  // An ID typed into a block only exists in the server it was made in.
  if (data.some((d) => /\b\d{17,20}\b/.test(d))) {
    res.push(
      "Has IDs of channels, roles or users of the server it was made in"
    );
  }
  return res;
}
