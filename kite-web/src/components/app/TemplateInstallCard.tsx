import { parameterDefaults, Template } from "@/lib/flow/templates";
import { cn } from "@/lib/utils";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import {
  useBundleInstallCreateMutation,
  useMessageInstanceCreateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useBundleInstalls, useCommands } from "@/lib/hooks/api";
import { track } from "@/lib/analytics";
import { CircleAlertIcon, CircleCheckIcon } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import LoadingButton from "../common/LoadingButton";
import MessageEditorDialog from "../message/MessageEditorDialog";
import { Label } from "../ui/label";
import { MessageData } from "@/lib/types/message.gen";
import {
  TemplateParameterFields,
  useTemplateGuild,
} from "./TemplateParameterFields";

// useExistingCommandNames are the names of the app's commands. Importing a
// command with a name the app already uses fails.
function useExistingCommandNames() {
  const commands = useCommands();
  return useMemo(
    () => new Set(commands?.flatMap((c) => (c ? [c.name] : []))),
    [commands]
  );
}

export function TemplateInstallCard({ template }: { template: Template }) {
  const appId = useAppId();
  const existingCommands = useExistingCommandNames();
  // An app has a template once, adding it again adds to that install.
  const existing = useBundleInstalls()?.find(
    (i) => i?.bundle_id === template.id
  );
  // The items it added that the app still has.
  const installedKeys = new Set(
    existing?.items.filter((i) => !i.deleted).map((i) => i.key)
  );
  const commandExists = (c: Template["commands"][number]) =>
    existingCommands.has(c.name) || installedKeys.has(c.key);
  const listenerExists = (l: Template["eventListeners"][number]) =>
    installedKeys.has(l.key);

  const bundleInstallMutation = useBundleInstallCreateMutation(appId);
  const postMutation = useMessageInstanceCreateMutation(appId);

  const [values, setValues] = useState(() =>
    parameterDefaults(template.parameters)
  );
  // The messages the user wrote in the message editor, by their key.
  const [customMessages, setCustomMessages] = useState<
    Record<string, MessageData>
  >({});

  // Channels and roles are picked from one server for all parameters.
  const [guildId, setGuildId] = useTemplateGuild();
  const [skipped, setSkipped] = useState<Set<string>>(new Set());
  const [installed, setInstalled] = useState<{
    commands: number;
    eventListeners: number;
    messages: number;
    // Messages that couldn't be posted, with the reason.
    unposted: { name: string; error: string }[];
  } | null>(null);

  const commands = template.commands.filter(
    (c) => !skipped.has(c.key) && !commandExists(c)
  );
  const eventListeners = template.eventListeners.filter(
    (l) => !skipped.has(l.key) && !listenerExists(l)
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
      (usedKeys.has(p.key) && !existing?.parameters?.[p.key])
  );
  const itemCount =
    commands.length +
    eventListeners.length +
    messages.length +
    variables.length;

  function toggle(key: string, included: boolean) {
    const next = new Set(skipped);
    if (included) next.delete(key);
    else next.add(key);
    setSkipped(next);
  }

  async function install() {
    const res = await bundleInstallMutation.mutateAsync({
      bundle_id: template.id,
      bundle_version: template.version,
      parameters: Object.fromEntries(
        parameters.flatMap((parameter) =>
          values[parameter.key] ? [[parameter.key, values[parameter.key]]] : []
        )
      ),
      items: [...commands, ...eventListeners, ...messages, ...variables].map(
        (i) => i.key
      ),
      messages: Object.fromEntries(
        customizable.flatMap((m) =>
          customMessages[m.key] ? [[m.key, customMessages[m.key]]] : []
        )
      ),
    });
    if (!res.success) {
      toast.error(
        `Failed to add ${template.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }

    const analytics = {
      source: "template",
      template: template.id,
      template_version: template.version,
    };
    // first is false when items are added to an install the app already has.
    track("template_added", {
      ...analytics,
      count: itemCount,
      first: !existing,
    });
    if (commands.length) {
      track("command_created", { ...analytics, count: commands.length });
    }
    if (eventListeners.length) {
      track("event_listener_created", {
        ...analytics,
        count: eventListeners.length,
      });
    }

    // Messages are posted after they were created, a message that can't be
    // posted can be posted again from this page.
    const unposted: { name: string; error: string }[] = [];
    for (const message of messages) {
      const channelId = message.sendTo && values[message.sendTo];
      const item = res.data.items.find(
        (i) => i.type === "message" && i.key === message.key
      );
      if (!channelId || !guildId || !item) continue;

      const postRes = await postMutation.mutateAsync({
        messageId: item.id,
        discord_guild_id: guildId,
        discord_channel_id: channelId,
      });
      if (postRes.success) {
        track("message_sent", analytics);
      } else {
        unposted.push({ name: message.name, error: postRes.error.message });
      }
    }

    setInstalled({
      commands: commands.length,
      eventListeners: eventListeners.length,
      messages: messages.length,
      unposted,
    });
  }

  if (installed) {
    return (
      <Card className="p-5">
        <div className="flex items-center gap-2 font-semibold mb-1">
          <CircleCheckIcon className="h-5 w-5 text-green-500" />
          Added to your app
        </div>
        <p className="text-sm text-muted-foreground mb-4">
          It&apos;s enabled and yours to change in the editor, like anything you
          built yourself.
        </p>
        {installed.unposted.map((m) => (
          <div
            key={m.name}
            className="flex gap-2 text-sm mb-4 text-amber-600 dark:text-amber-500"
          >
            <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
            <span>
              Couldn&apos;t post the {m.name} message: {m.error}. Post it again
              from this page once that&apos;s fixed.
            </span>
          </div>
        ))}
        <div className="flex flex-col gap-2">
          {!!installed.commands && (
            <Button variant="outline" asChild>
              <Link
                href={{ pathname: "/apps/[appId]/commands", query: { appId } }}
              >
                View commands
              </Link>
            </Button>
          )}
          {!!installed.eventListeners && (
            <Button variant="outline" asChild>
              <Link
                href={{
                  pathname: eventSourceName("discord").path,
                  query: { appId },
                }}
              >
                View event listeners
              </Link>
            </Button>
          )}
          {!!installed.messages && (
            <Button variant="outline" asChild>
              <Link
                href={{ pathname: "/apps/[appId]/messages", query: { appId } }}
              >
                View message templates
              </Link>
            </Button>
          )}
        </div>
      </Card>
    );
  }

  // Items of a newer version can't be added to an older install.
  if (existing && existing.bundle_version < template.version) {
    return (
      <Card className="p-5">
        <div className="font-semibold mb-1">Add to your app</div>
        <p className="text-sm text-muted-foreground">
          Your app has version {existing.bundle_version} of this template.
          Update it to version {template.version} first, then you can add what
          it doesn&apos;t have yet.
        </p>
      </Card>
    );
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-4">Add to your app</div>

      {!!parameters.length && (
        <div className="mb-5">
          <TemplateParameterFields
            parameters={parameters}
            values={values}
            onChange={setValues}
            guildId={guildId}
            onGuildChange={setGuildId}
          />
        </div>
      )}

      {customizable.map((message) => (
        <div key={message.key} className="space-y-1.5 mb-5">
          <Label>{message.name} Message</Label>
          {message.description && (
            <div className="text-xs text-muted-foreground">
              {message.description}
            </div>
          )}
          <MessageEditorDialog
            message={customMessages[message.key] ?? message.data}
            onClose={(data, changed) => {
              // Unchanged, it stays the template's, which updates replace.
              if (!changed) return;
              setCustomMessages((prev) => ({ ...prev, [message.key]: data }));
            }}
          >
            <Button variant="outline" className="w-full">
              {customMessages[message.key] ? "Edit message" : "Write message"}
            </Button>
          </MessageEditorDialog>
        </div>
      ))}

      <div className="text-sm font-medium mb-2.5">Include</div>
      <div className="space-y-2.5 mb-5">
        {template.commands.map((command) => (
          <ItemRow
            key={command.key}
            exists={commandExists(command)}
            included={!skipped.has(command.key)}
            onToggle={(included) => toggle(command.key, included)}
          >
            /{command.name}
          </ItemRow>
        ))}
        {template.eventListeners.map((listener) => (
          <ItemRow
            key={listener.key}
            exists={listenerExists(listener)}
            included={!skipped.has(listener.key)}
            onToggle={(included) => toggle(listener.key, included)}
          >
            {discordEventType(listener.type)?.label ?? listener.type}
          </ItemRow>
        ))}
        {template.messages.map((message) => {
          const exists = installedKeys.has(message.key);
          return (
            <ItemRow key={message.key} exists={exists}>
              {message.name} message
              {message.sendTo && !exists && (
                <span className="text-muted-foreground">
                  , posted in the channel
                </span>
              )}
            </ItemRow>
          );
        })}
        {template.variables.map((variable) => (
          <ItemRow key={variable.key} exists={installedKeys.has(variable.key)}>
            {variable.name} stored variable
          </ItemRow>
        ))}
      </div>

      {itemCount === 0 ? (
        <Button className="w-full" disabled>
          Nothing selected
        </Button>
      ) : (
        <LoadingButton
          className="w-full"
          onClick={install}
          loading={bundleInstallMutation.isPending}
        >
          Add {itemCount} {itemCount === 1 ? "item" : "items"}
        </LoadingButton>
      )}
    </Card>
  );
}

// ItemRow is an item of the template to include. Without onToggle it's always
// included, unless the app has it already.
function ItemRow({
  exists,
  included = true,
  onToggle,
  children,
}: {
  exists: boolean;
  included?: boolean;
  onToggle?: (included: boolean) => void;
  children: React.ReactNode;
}) {
  return (
    <label
      className={cn(
        "flex items-center gap-2.5 text-sm",
        onToggle && "cursor-pointer"
      )}
    >
      <Checkbox
        checked={!exists && included}
        disabled={exists || !onToggle}
        onCheckedChange={(checked) => onToggle?.(!!checked)}
      />
      <span className={exists ? "text-muted-foreground" : ""}>{children}</span>
      {exists && (
        <span className="text-xs text-muted-foreground ml-auto">
          already in your app
        </span>
      )}
    </label>
  );
}
