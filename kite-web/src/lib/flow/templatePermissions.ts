import { getBlockDefinition } from "../blocks";
import {
  decodePermissionsBitset,
  PermissionLabel,
  permissionBits,
} from "../discord/permissions";
import { FlowData, FlowNode } from "../types/flow.gen";
import { isNoteNode } from "./notes";
import { Template } from "./templates";

// Blocks that modify server members or roles, which Discord only allows when
// the bot's highest role is above the member or role it acts on.
const roleHierarchyTypes = new Set([
  "action_member_ban",
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
}

// nodePermissions returns the Discord permissions required by a single flow node.
function nodePermissions(node: FlowNode): PermissionLabel[] {
  if (!node.type || isNoteNode(node)) return [];
  const def = getBlockDefinition(node.type);
  if (!def?.permissions) return [];
  return typeof def.permissions === "function"
    ? def.permissions(node.data ?? {})
    : def.permissions;
}

// flowPermissions derives the Discord permissions required by a single flow.
export function flowPermissions(flow: FlowData): PermissionLabel[] {
  const permissions = new Set<PermissionLabel>();
  for (const node of flow.nodes ?? []) {
    for (const perm of nodePermissions(node)) {
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
      if (node.type && roleHierarchyTypes.has(node.type)) {
        needsRoleHierarchy = true;
      }
      for (const perm of nodePermissions(node)) {
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
      const msgData = message.data as any;
      if (msgData?.embeds?.length) permissions.add("Embed Links");
      if (msgData?.attachments?.length) permissions.add("Attach Files");
    }
    if (message.flowSources) {
      for (const flow of Object.values(message.flowSources)) {
        checkNodes(flow.nodes ?? []);
      }
    }
  }

  return {
    // In the order Discord lists them.
    permissions: permissionBits
      .map((p) => p.label)
      .filter((label) => permissions.has(label)),
    needsRoleHierarchy,
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
    parts.push(`${new Intl.ListFormat("en").format(derived.permissions)}.`);
  }
  if (derived.needsRoleHierarchy) {
    parts.push(
      "Keep the bot's role above the members and roles it manages, since that's how Discord decides who a bot can act on."
    );
  }

  return {
    question: "What permissions does the bot need?",
    answer: parts.join(" "),
  };
}
