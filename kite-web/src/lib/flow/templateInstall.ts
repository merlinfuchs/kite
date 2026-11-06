import { MessageData } from "../types/message.gen";
import { BundleInstall, BundleInstallCreateRequest } from "../types/wire.gen";
import { eventListenerLabel } from "./eventTypes";
import { Template } from "./templates";

// installSelection is what adding a template adds to an app: the items that
// aren't skipped and that the app doesn't have yet, and the parameters to ask
// for. install is the app's install of the template, if it has one. Commands
// whose names the app uses already are added too, with a number.
export function installSelection(
  template: Template,
  install: BundleInstall | undefined,
  skipped: Set<string>
) {
  // The items it added that the app still has.
  const installedKeys = new Set(
    install?.items.filter((i) => !i.deleted).map((i) => i.key)
  );
  const commands = template.commands.filter(
    (c) => !skipped.has(c.key) && !installedKeys.has(c.key)
  );
  const eventListeners = template.eventListeners.filter(
    (l) => !skipped.has(l.key) && !installedKeys.has(l.key)
  );
  // Messages and variables are always added, the commands and event listeners
  // use them.
  const messages = template.messages.filter((m) => !installedKeys.has(m.key));
  const customizable = messages.filter((m) => m.customize);
  const variables = template.variables.filter((v) => !installedKeys.has(v.key));
  // Only the parameters of what's added are asked for, and not the ones the
  // template already has values for. The channels messages are posted in are
  // asked for again, they're picked from a server.
  const usedKeys = new Set(
    [...commands, ...eventListeners, ...messages].flatMap(
      (i) => i.parameterKeys
    )
  );
  const postTo = new Set(messages.flatMap((m) => (m.sendTo ? [m.sendTo] : [])));
  const parameters = template.parameters.filter(
    (p) =>
      postTo.has(p.key) ||
      (usedKeys.has(p.key) && !install?.parameters?.[p.key])
  );

  return {
    installedKeys,
    commands,
    eventListeners,
    messages,
    customizable,
    variables,
    parameters,
  };
}

export type InstallSelection = ReturnType<typeof installSelection>;

// installRequest adds the selected items with the values of the parameters
// and the messages the user wrote, by their keys.
export function installRequest(
  template: Template,
  selection: InstallSelection,
  values: Record<string, string>,
  customMessages: Record<string, MessageData>
): BundleInstallCreateRequest {
  const { commands, eventListeners, messages, variables } = selection;
  return {
    bundle_id: template.id,
    bundle_version: template.version,
    parameters: Object.fromEntries(
      selection.parameters.flatMap((parameter) =>
        values[parameter.key] ? [[parameter.key, values[parameter.key]]] : []
      )
    ),
    items: [...commands, ...eventListeners, ...messages, ...variables].map(
      (i) => i.key
    ),
    messages: Object.fromEntries(
      selection.customizable.flatMap((m) =>
        customMessages[m.key] ? [[m.key, customMessages[m.key]]] : []
      )
    ),
  };
}

// updateSelection is what updating an install to the template's version
// changes: unedited items are replaced, edited ones only if they're in
// overwrite, and the parameters the replaced ones need that the install has
// no value for are asked for.
export function updateSelection(
  template: Template,
  install: BundleInstall,
  overwrite: Set<string>
) {
  const items = install.items.filter((item) => !item.deleted);
  const outdated = install.bundle_version < template.version;
  // Items that are no longer in the template stay as they are, and so do
  // variables, which only have a name.
  const updatable = outdated
    ? items.filter(
        (item) => item.type !== "variable" && templateItem(template, item.key)
      )
    : [];
  const edited = updatable.filter((item) => item.edited);
  const unedited = updatable.filter((item) => !item.edited);
  const replaced = updatable.filter(
    (item) => !item.edited || overwrite.has(item.key)
  );

  // Parameters the new version uses that the install has no value for.
  const usedKeys = new Set(
    replaced.flatMap((item) => templateItem(template, item.key)!.parameterKeys)
  );
  // The ones its flows have already, even without a value, aren't new.
  const existingKeys = new Set(
    items.flatMap((item) => Object.keys(item.parameters ?? {}))
  );
  const missingParameters = template.parameters.filter(
    (s) =>
      usedKeys.has(s.key) &&
      !existingKeys.has(s.key) &&
      !install.parameters?.[s.key]
  );

  return { items, outdated, updatable, edited, unedited, missingParameters };
}

export function templateItem(template: Template, key: string) {
  const command = template.commands.find((c) => c.key === key);
  if (command) return { label: `/${command.name}`, ...command };
  const listener = template.eventListeners.find((l) => l.key === key);
  if (listener) {
    return {
      label: eventListenerLabel(listener.source, listener.type),
      ...listener,
    };
  }
  const variable = template.variables.find((v) => v.key === key);
  if (variable) {
    return {
      label: `${variable.name} stored variable`,
      ...variable,
      parameterKeys: [],
    };
  }
  const message = template.messages.find((m) => m.key === key);
  if (message) {
    return {
      label: `${message.name} message`,
      ...message,
    };
  }
}

export function itemLabel(template: Template, key: string) {
  return templateItem(template, key)?.label ?? key;
}
