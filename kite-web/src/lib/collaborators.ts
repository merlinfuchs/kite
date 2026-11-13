// Roles of collaborators of an app, see AppCollaboratorRole in the service.
export type CollaboratorRole = "owner" | "admin" | "developer";

export const collaboratorRoles: Record<
  CollaboratorRole,
  { label: string; description: string }
> = {
  owner: {
    label: "Owner",
    description: "Can do everything, including deleting the app.",
  },
  admin: {
    label: "Admin",
    description:
      "Can edit everything, change the app's settings, and invite developers.",
  },
  developer: {
    label: "Developer",
    description:
      "Can edit commands, event listeners, message templates, and everything else, but not the app's settings.",
  },
};

export function collaboratorRoleLabel(role: string | undefined) {
  return collaboratorRoles[role as CollaboratorRole]?.label ?? role ?? "";
}

// The roles a collaborator with the role can invite. Only the owner picks
// admins.
export function invitableRoles(role: string | undefined): CollaboratorRole[] {
  if (role === "owner") return ["admin", "developer"];
  if (role === "admin") return ["developer"];
  return [];
}

// Whether a collaborator with the role can remove one with the other role.
// Admins can't remove other admins or themselves, they leave the app instead.
export function canRemoveRole(role: string | undefined, other: string) {
  return invitableRoles(role).includes(other as CollaboratorRole);
}

export function canManageSettings(role: string | undefined) {
  return role === "owner" || role === "admin";
}

// What the actions of the audit log say, after the name of the user.
const auditActions: Record<string, string> = {
  "app.update": "updated the app",
  "app.status_update": "changed the status of the app",
  "app.credit_notices_update": "changed the credit notices",
  "app.token_update": "changed the bot token",
  "collaborator.invite": "invited",
  "collaborator.invite_revoke": "revoked the invite of",
  "collaborator.join": "accepted the invite and joined",
  "collaborator.decline": "declined the invite",
  "collaborator.remove": "removed the collaborator",
  "collaborator.leave": "left the app",
  "collaborator.role_update": "changed the role of",
  "share_code.create": "created a share code",
  "billing.plan_update": "changed the plan",
  "command.create": "created the command",
  "command.import": "imported commands",
  "command.deploy": "deployed the commands",
  "command.update": "edited the command",
  "command.delete": "deleted the command",
  "command.enable": "enabled the command",
  "command.disable": "disabled the command",
  "event_listener.create": "created the event listener",
  "event_listener.import": "imported event listeners",
  "event_listener.update": "edited the event listener",
  "event_listener.delete": "deleted the event listener",
  "event_listener.enable": "enabled the event listener",
  "event_listener.disable": "disabled the event listener",
  "event_listener.webhook_secret_update": "regenerated the webhook URL of",
  "plugin.create": "added the plugin",
  "plugin.update": "changed the plugin",
  "plugin.delete": "removed the plugin",
  "plugin.enable": "enabled the plugin",
  "plugin.disable": "disabled the plugin",
  "variable.create": "created the stored variable",
  "variable.import": "imported stored variables",
  "variable.update": "edited the stored variable",
  "variable.delete": "deleted the stored variable",
  "variable.value_set": "changed a stored value of",
  "variable.value_delete": "deleted a stored value of",
  "secret.create": "created the secret",
  "secret.update": "changed the secret",
  "secret.delete": "deleted a secret",
  "integration.connect": "connected the integration",
  "integration.update": "changed the integration",
  "integration.remove": "removed the integration",
  "message.create": "created the message template",
  "message.import": "imported message templates",
  "message.update": "edited the message template",
  "message.delete": "deleted the message template",
  "message.send": "sent the message template",
  "message.sent_update": "updated a sent copy of",
  "message.sent_delete": "deleted a sent copy of",
  "template.install": "added the template",
  "template.update": "updated the template",
  "template.uninstall": "removed a template",
  "template.parameters_update": "changed the parameters of a template",
  "bundle.import": "imported a template",
  "bundle.create": "shared the template",
  "bundle.version_create": "shared a new version of a template",
  "bot.profile_update": "changed the bot's profile",
  "guild.leave": "made the bot leave a server",
  "guild.profile_update": "changed the bot's profile in a server",
};

export function auditActionLabel(action: string) {
  return auditActions[action] ?? action.replace(/[._]/g, " ");
}

// The kinds of changes the audit log can be filtered by, with the part of
// their actions before the dot.
export const auditCategories: {
  value: string;
  label: string;
  kinds: string[];
}[] = [
  { value: "commands", label: "Commands", kinds: ["command"] },
  {
    value: "event_listeners",
    label: "Event listeners",
    kinds: ["event_listener"],
  },
  { value: "messages", label: "Message templates", kinds: ["message"] },
  { value: "variables", label: "Stored variables", kinds: ["variable"] },
  { value: "secrets", label: "Secrets", kinds: ["secret"] },
  { value: "integrations", label: "Integrations", kinds: ["integration"] },
  { value: "templates", label: "Templates", kinds: ["template", "bundle"] },
  { value: "plugins", label: "Plugins", kinds: ["plugin"] },
  { value: "servers", label: "Servers", kinds: ["guild"] },
  { value: "collaborators", label: "Collaborators", kinds: ["collaborator"] },
  {
    value: "settings",
    label: "Settings",
    kinds: ["app", "billing", "bot", "share_code"],
  },
];
