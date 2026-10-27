import { getBlockDefinition } from "../blocks";
import {
  decodePermissionsBitset,
  PermissionLabel,
} from "../discord/permissions";
import { FlowData, FlowNode } from "../types/flow.gen";
import { isNoteNode } from "./notes";
import { Template } from "./templates";

// Blocks that modify server members or roles, which Discord only allows when
// the bot's highest role is above the member or role it acts on.
const roleHierarchyTypes = new Set([
  "action_member_ban",
  "action_member_unban",
  "action_member_kick",
  "action_member_timeout",
  "action_member_edit",
  "action_member_role_add",
  "action_member_role_remove",
  "action_role_edit",
  "action_role_delete",
]);

export interface TemplatePermissions {
  // Distinct permissions required by the template's blocks and messages.
  permissions: PermissionLabel[];
  // Whether the bot's role must be above the members and roles it acts on.
  needsRoleHierarchy: boolean;
  // Privileged intents the template needs turned on in the developer portal.
  intents: string[];
}

// flowPermissions derives the Discord permissions required by a single flow.
export function flowPermissions(flow: FlowData): PermissionLabel[] {
  const permissions = new Set<PermissionLabel>();
  for (const node of flow.nodes ?? []) {
    if (!node.type || isNoteNode(node)) continue;
    const def = getBlockDefinition(node.type);
    if (!def?.permissions) continue;
    const perms =
      typeof def.permissions === "function"
        ? def.permissions(node.data ?? {})
        : def.permissions;
    for (const perm of perms) {
      permissions.add(perm);
    }
  }
  return Array.from(permissions);
}

// templatePermissions derives the Discord permissions and setup requirements
// of a template or a subset of its items (e.g. while installing).
export function templatePermissions(
  template: Template,
  includedKeys?: Set<string>
): TemplatePermissions {
  const permissions = new Set<PermissionLabel>();
  let needsRoleHierarchy = false;

  const checkNodes = (nodes: FlowNode[]) => {
    for (const node of nodes) {
      if (!node.type || isNoteNode(node)) continue;
      if (roleHierarchyTypes.has(node.type)) {
        needsRoleHierarchy = true;
      }
      const def = getBlockDefinition(node.type);
      if (!def?.permissions) continue;
      const perms =
        typeof def.permissions === "function"
          ? def.permissions(node.data ?? {})
          : def.permissions;
      for (const perm of perms) {
        permissions.add(perm);
      }
    }
  };

  const commands = includedKeys
    ? template.commands.filter((c) => includedKeys.has(c.key))
    : template.commands;
  for (const command of commands) {
    checkNodes(command.flowSource.nodes ?? []);
  }

  const eventListeners = includedKeys
    ? template.eventListeners.filter((l) => includedKeys.has(l.key))
    : template.eventListeners;
  for (const listener of eventListeners) {
    checkNodes(listener.flowSource.nodes ?? []);
  }

  const messages = includedKeys
    ? template.messages.filter((m) => includedKeys.has(m.key))
    : template.messages;
  for (const message of messages) {
    if (message.sendTo) {
      permissions.add("View Channel");
      permissions.add("Send Messages");
      const embeds = (message.data as any)?.embeds;
      if (Array.isArray(embeds) && embeds.length > 0) {
        permissions.add("Embed Links");
      }
      const attachments = (message.data as any)?.attachments;
      if (Array.isArray(attachments) && attachments.length > 0) {
        permissions.add("Attach Files");
      }
    }
    if (message.flowSources) {
      for (const flow of Object.values(message.flowSources)) {
        checkNodes(flow.nodes ?? []);
      }
    }
  }

  // Privileged intents
  const intents = new Set<string>();
  for (const listener of eventListeners) {
    if (
      listener.type === "guild_member_add" ||
      listener.type === "guild_member_remove" ||
      listener.type === "guild_member_update"
    ) {
      intents.add("Server Members Intent");
    }
    if (
      listener.type === "message_create" ||
      listener.type === "message_update"
    ) {
      intents.add("Message Content Intent");
    }
  }

  return {
    permissions: Array.from(permissions),
    needsRoleHierarchy,
    intents: Array.from(intents),
  };
}

// missingPermissions returns any required permissions that the bot does not have
// in the selected guild bitset. Administrator grants all permissions.
export function missingPermissions(
  required: PermissionLabel[],
  botPermissionsBitset?: string | null
): PermissionLabel[] {
  if (!botPermissionsBitset) return [];
  const botPerms = decodePermissionsBitset(botPermissionsBitset);
  if (botPerms.some((p) => p.label === "Administrator")) {
    return [];
  }
  const botSet = new Set(botPerms.map((p) => p.label));
  return required.filter((p) => !botSet.has(p));
}

// templatePermissionFaq generates the FAQ question & answer for a template's
// required permissions.
export function templatePermissionFaq(
  template: Template
): { question: string; answer: string } | null {
  const derived = templatePermissions(template);
  if (derived.permissions.length === 0 && !derived.needsRoleHierarchy) {
    return null;
  }

  const parts: string[] = [];
  if (derived.permissions.length > 0) {
    const list = formatPermissionList(derived.permissions);
    parts.push(`${list} in the channels it runs in.`);
  }

  if (derived.needsRoleHierarchy) {
    parts.push(
      "Keep the bot's role above the members and roles it manages, since that's how Discord decides who a bot can act on."
    );
  }

  if (derived.intents.includes("Message Content Intent")) {
    parts.push(
      "Turn on the Message Content Intent under Bot in the Discord Developer Portal, or the bot can't read message content."
    );
  }
  if (derived.intents.includes("Server Members Intent")) {
    parts.push(
      "Turn on the Server Members Intent under Bot in the Discord Developer Portal."
    );
  }

  return {
    question: "What permissions does the bot need?",
    answer: parts.join(" "),
  };
}

function formatPermissionList(permissions: PermissionLabel[]): string {
  if (permissions.length === 0) return "";
  if (permissions.length === 1) return permissions[0];
  if (permissions.length === 2)
    return `${permissions[0]} and ${permissions[1]}`;
  return `${permissions.slice(0, -1).join(", ")} and ${
    permissions[permissions.length - 1]
  }`;
}
