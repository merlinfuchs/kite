import { getBlockDefinition } from "../blocks";
import {
  decodePermissionsBitset,
  PermissionFlag,
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
  permissions: PermissionFlag[];
  // Whether the bot's role must be above the members and roles it acts on.
  needsRoleHierarchy: boolean;
}

// nodePermissions returns the Discord permissions required by a single flow node.
function nodePermissions(node: FlowNode): PermissionFlag[] {
  if (!node.type || isNoteNode(node)) return [];
  const def = getBlockDefinition(node.type);
  if (!def?.permissions) return [];
  return typeof def.permissions === "function"
    ? def.permissions(node.data ?? {})
    : def.permissions;
}

// flowPermissions derives the Discord permissions required by a single flow.
export function flowPermissions(flow: FlowData): PermissionFlag[] {
  const permissions = new Set<PermissionFlag>();
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
  const permissions = new Set<PermissionFlag>();
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
      permissions.add("VIEW_CHANNEL");
      permissions.add("SEND_MESSAGES");
      const msgData = message.data as any;
      if (msgData?.embeds?.length) permissions.add("EMBED_LINKS");
      if (msgData?.attachments?.length) permissions.add("ATTACH_FILES");
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
      .map((p) => p.flag)
      .filter((flag) => permissions.has(flag)),
    needsRoleHierarchy,
  };
}

// missingPermissions returns any required permissions that the bot does not have
// in the selected guild bitset. Administrator grants all permissions.
export function missingPermissions(
  required: PermissionFlag[],
  botPermissionsBitset?: string | null
): PermissionFlag[] {
  if (!botPermissionsBitset) return [];
  const botPerms = decodePermissionsBitset(botPermissionsBitset);
  if (botPerms.some((p) => p.flag === "ADMINISTRATOR")) {
    return [];
  }
  const botSet = new Set(botPerms.map((p) => p.flag));
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

  // Kite's invite link adds bots with Administrator, which covers all of them
  // but not the role hierarchy.
  const parts: string[] = [];
  if (derived.permissions.length > 0) {
    parts.push(
      `${formatList(
        derived.permissions.map(permissionLabel)
      )}. Kite's invite link gives the bot Administrator by default, which includes them.`
    );
  }
  if (derived.needsRoleHierarchy) {
    parts.push(
      "Administrator doesn't get around Discord's role hierarchy, so keep the bot's role above the members and roles it acts on."
    );
  }

  return {
    question: "What permissions does the bot need?",
    answer: parts.join(" "),
  };
}

// permissionLabel is the name Discord shows for the permission.
export function permissionLabel(flag: PermissionFlag) {
  return permissionBits.find((p) => p.flag === flag)!.label;
}

// formatList joins the items like the templates' own text, without a comma
// before the "and".
function formatList(items: string[]) {
  if (items.length <= 1) return items.join("");
  return `${items.slice(0, -1).join(", ")} and ${items[items.length - 1]}`;
}
