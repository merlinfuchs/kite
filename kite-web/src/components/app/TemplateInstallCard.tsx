import { parameterDefaults, Template } from "@/lib/flow/templates";
import { installRequest, installSelection } from "@/lib/flow/templateInstall";
import { BundleInstall } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import {
  useBundleInstallCreateMutation,
  useMessageInstanceCreateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { PermissionFlag } from "@/lib/discord/permissions";
import {
  useAppStateGuilds,
  useBundleInstalls,
  useCommands,
} from "@/lib/hooks/api";
import { track } from "@/lib/analytics";
import {
  CircleAlertIcon,
  CircleCheckIcon,
  PlugIcon,
  ShieldAlertIcon,
} from "lucide-react";
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
import { useMissingFlowIntegrations } from "@/lib/integrations/hooks";
import { Integration } from "@/lib/integrations/types";
import { AppIntegrationEnableButton } from "./AppIntegrationList";
import {
  missingPermissions,
  permissionLabel,
  templatePermissions,
} from "@/lib/flow/templatePermissions";
import {
  TemplateParameterFields,
  useTemplateGuild,
} from "./TemplateParameterFields";
import { PlanLimitNotice, useExceededLimit } from "./PlanLimit";

// useCommandNames are the names of the app's commands. An added command with
// one of them gets a number.
function useCommandNames() {
  const commands = useCommands();
  return useMemo(
    () => new Set(commands?.flatMap((c) => (c ? [c.name] : []))),
    [commands]
  );
}

export function TemplateInstallCard({
  template,
  from,
}: {
  template: Template;
  // Where the template is added from, like "flow_ai", for analytics.
  from?: string;
}) {
  const appId = useAppId();
  const commandNames = useCommandNames();
  // An app has a template once, adding it again adds to that install.
  const existing = useBundleInstalls()?.find(
    (i): i is BundleInstall => i?.bundle_id === template.id
  );
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

  const selection = installSelection(template, existing, skipped);
  const {
    installedKeys,
    commands,
    eventListeners,
    messages,
    variables,
    customizable,
    parameters,
  } = selection;
  // Integrations the blocks of what's added need, like ER:LC, which can be
  // enabled right here. Without them the blocks fail when they run. Once
  // added, it's what was added.
  const added = installed ? template : { commands, eventListeners, messages };
  const missingIntegrations = useMissingFlowIntegrations([
    ...[...added.commands, ...added.eventListeners].map((i) => i.flowSource),
    ...added.messages.flatMap((m) => Object.values(m.flowSources)),
  ]);
  const itemCount =
    commands.length +
    eventListeners.length +
    messages.length +
    variables.length;

  const exceeded = useExceededLimit({
    commands: commands.length,
    eventListenerSources: eventListeners.map((l) => l.source),
    messages: messages.length,
    variables: variables.length,
  });

  const guilds = useAppStateGuilds();
  const includedKeys = useMemo(
    () =>
      new Set([
        ...commands.map((c) => c.key),
        ...eventListeners.map((l) => l.key),
        ...messages.map((m) => m.key),
      ]),
    [commands, eventListeners, messages]
  );
  const derived = useMemo(
    () => templatePermissions(template, includedKeys),
    [template, includedKeys]
  );
  // Until a server is picked, every server the bot is in is checked.
  const missing = useMemo(() => {
    const checked = (guilds ?? []).filter((g) => !guildId || g?.id === guildId);
    const permissions = new Set<PermissionFlag>();
    const guildNames: string[] = [];
    for (const guild of checked) {
      const perms = missingPermissions(
        derived.permissions,
        guild?.bot_permissions
      );
      if (perms.length === 0) continue;
      perms.forEach((p) => permissions.add(p));
      guildNames.push(guild!.name);
    }
    return { permissions: Array.from(permissions), guildNames };
  }, [derived.permissions, guilds, guildId]);

  function toggle(key: string, included: boolean) {
    const next = new Set(skipped);
    if (included) next.delete(key);
    else next.add(key);
    setSkipped(next);
  }

  async function install() {
    const res = await bundleInstallMutation.mutateAsync(
      installRequest(template, selection, values, customMessages)
    );
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
      ...(from && { from }),
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
        {missingIntegrations.map((integration) => (
          <MissingIntegration key={integration.id} integration={integration} />
        ))}
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
            exists={installedKeys.has(command.key)}
            included={!skipped.has(command.key)}
            onToggle={(included) => toggle(command.key, included)}
            note={
              !skipped.has(command.key) && commandNames.has(command.name)
                ? "name taken, gets a number"
                : undefined
            }
          >
            /{command.name}
          </ItemRow>
        ))}
        {template.eventListeners.map((listener) => (
          <ItemRow
            key={listener.key}
            exists={installedKeys.has(listener.key)}
            included={!skipped.has(listener.key)}
            onToggle={(included) => toggle(listener.key, included)}
            description={listener.description}
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

      {missingIntegrations.map((integration) => (
        <MissingIntegration key={integration.id} integration={integration} />
      ))}

      {missing.permissions.length > 0 && (
        <div className="flex gap-2.5 rounded-lg border border-amber-500/50 bg-amber-500/10 p-3 text-xs text-amber-600 dark:text-amber-500 mb-5">
          <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
          <div>
            <span className="font-semibold">Missing permissions: </span>
            Your bot needs the{" "}
            {new Intl.ListFormat("en").format(
              missing.permissions.map(permissionLabel)
            )}{" "}
            {missing.permissions.length === 1 ? "permission" : "permissions"}{" "}
            {missing.guildNames.length > 3
              ? `in ${missing.guildNames.length} of your servers`
              : `in ${new Intl.ListFormat("en").format(missing.guildNames)}`}
            .
          </div>
        </div>
      )}

      {derived.needsRoleHierarchy && (
        <div className="flex gap-2.5 rounded-lg border border-border bg-muted/40 p-3 text-xs text-muted-foreground mb-5">
          <ShieldAlertIcon className="h-4 w-4 flex-none mt-0.5 text-muted-foreground" />
          <div>
            <span className="font-semibold text-foreground">
              Role hierarchy:{" "}
            </span>
            Put your bot&apos;s role above the roles of the members it acts on
            so it can manage them.
          </div>
        </div>
      )}

      {exceeded && (
        <PlanLimitNotice exceeded={exceeded} from="template" className="mb-5" />
      )}

      {itemCount === 0 ? (
        <Button className="w-full" disabled>
          Nothing selected
        </Button>
      ) : (
        <LoadingButton
          className="w-full"
          onClick={install}
          disabled={!!exceeded}
          loading={bundleInstallMutation.isPending}
        >
          Add {itemCount} {itemCount === 1 ? "item" : "items"}
        </LoadingButton>
      )}
    </Card>
  );
}

function MissingIntegration({ integration }: { integration: Integration }) {
  return (
    <div className="flex items-center gap-3 rounded-md border p-3 mb-5 text-sm">
      <PlugIcon className="h-4 w-4 flex-none text-muted-foreground" />
      <div className="flex-1 text-muted-foreground">
        Uses {integration.name}, which isn&apos;t enabled in your app. Its
        blocks fail until you enable it.
      </div>
      <AppIntegrationEnableButton integration={integration} />
    </div>
  );
}

// ItemRow is an item of the template to include. Without onToggle it's always
// included, unless the app has it already.
function ItemRow({
  exists,
  included = true,
  onToggle,
  description,
  note,
  children,
}: {
  exists: boolean;
  included?: boolean;
  onToggle?: (included: boolean) => void;
  description?: string;
  // Shown next to it, like what adding it changes.
  note?: string;
  children: React.ReactNode;
}) {
  return (
    <label
      className={cn(
        "flex items-start gap-2.5 text-sm",
        onToggle && "cursor-pointer"
      )}
    >
      <Checkbox
        className="mt-0.5"
        checked={!exists && included}
        disabled={exists || !onToggle}
        onCheckedChange={(checked) => onToggle?.(!!checked)}
      />
      <div className="min-w-0 flex-1">
        <div className="flex gap-2">
          <span className={exists ? "text-muted-foreground" : ""}>
            {children}
          </span>
          {(exists || note) && (
            <span className="text-xs leading-5 text-muted-foreground ml-auto whitespace-nowrap">
              {exists ? "already in your app" : note}
            </span>
          )}
        </div>
        {description && (
          <div className="text-xs text-muted-foreground">{description}</div>
        )}
      </div>
    </label>
  );
}
