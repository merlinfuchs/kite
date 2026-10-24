import { Template, TemplateSetting } from "@/lib/flow/templates";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import { useBundleInstallCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import {
  useAppStateGuilds,
  useBundleInstalls,
  useCommands,
  useEventListeners,
} from "@/lib/hooks/api";
import { track } from "@/lib/analytics";
import { CircleCheckIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import ChannelSelect from "../common/ChannelSelect";
import GuildSelect from "../common/GuildSelect";
import RoleSelect from "../common/RoleSelect";

// useExistingCommandNames are the names of the app's commands. Importing a
// command with a name the app already uses fails.
function useExistingCommandNames() {
  const commands = useCommands();
  return useMemo(
    () => new Set(commands?.flatMap((c) => (c ? [c.name] : []))),
    [commands]
  );
}

// useInstalledItemKeys are the keys of the template's items that an earlier
// install added and the app still has.
function useInstalledItemKeys(templateId: string) {
  const installs = useBundleInstalls();
  const commands = useCommands();
  const eventListeners = useEventListeners();
  return useMemo(() => {
    const ids = new Set(
      [...(commands ?? []), ...(eventListeners ?? [])].flatMap((i) =>
        i ? [i.id] : []
      )
    );
    return new Set(
      installs?.flatMap((install) =>
        install?.bundle_id === templateId
          ? install.items.filter((i) => ids.has(i.id)).map((i) => i.key)
          : []
      )
    );
  }, [installs, commands, eventListeners, templateId]);
}

export function TemplateInstallCard({ template }: { template: Template }) {
  const appId = useAppId();
  const existingCommands = useExistingCommandNames();
  const installedKeys = useInstalledItemKeys(template.id);
  const commandExists = (c: Template["commands"][number]) =>
    existingCommands.has(c.name) || installedKeys.has(c.key);
  const listenerExists = (l: Template["eventListeners"][number]) =>
    installedKeys.has(l.key);

  const bundleInstallMutation = useBundleInstallCreateMutation(appId);

  const [values, setValues] = useState<Record<string, string>>({});

  // Channels and roles are picked from one server for all settings.
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useState<string | null>(null);
  const needsGuild = template.settings.some((s) =>
    pickerTypes.includes(s.type)
  );
  // Most apps are in a single server.
  useEffect(() => {
    if (!guildId && guilds?.length === 1) setGuildId(guilds[0]!.id);
  }, [guildId, guilds]);
  const [skipped, setSkipped] = useState<Set<string>>(new Set());
  const [installed, setInstalled] = useState<{
    commands: number;
    eventListeners: number;
  } | null>(null);

  const commands = template.commands.filter(
    (c) => !skipped.has(c.key) && !commandExists(c)
  );
  const eventListeners = template.eventListeners.filter(
    (l) => !skipped.has(l.key) && !listenerExists(l)
  );
  const itemCount = commands.length + eventListeners.length;

  function toggle(key: string, included: boolean) {
    const next = new Set(skipped);
    if (included) next.delete(key);
    else next.add(key);
    setSkipped(next);
  }

  async function install() {
    for (const setting of template.settings) {
      if (setting.required && !values[setting.key]?.trim()) {
        toast.error(`${setting.label} is required`);
        return;
      }
    }

    const res = await bundleInstallMutation.mutateAsync({
      bundle_id: template.id,
      bundle_version: template.version,
      name: template.name,
      description: template.description,
      settings: Object.fromEntries(
        template.settings.flatMap((setting) =>
          values[setting.key] ? [[setting.key, values[setting.key]]] : []
        )
      ),
      commands: commands.map((c) => ({
        key: c.key,
        command: {
          enabled: true,
          flow_source: c.flowSource(values),
        },
      })),
      event_listeners: eventListeners.map((l) => ({
        key: l.key,
        event_listener: {
          enabled: true,
          source: l.source,
          flow_source: l.flowSource(values),
        },
      })),
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
    if (commands.length) {
      track("command_created", { ...analytics, count: commands.length });
    }
    if (eventListeners.length) {
      track("event_listener_created", {
        ...analytics,
        count: eventListeners.length,
      });
    }

    setInstalled({
      commands: commands.length,
      eventListeners: eventListeners.length,
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
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-4">Add to your app</div>

      {!!template.settings.length && (
        <div className="space-y-4 mb-5">
          {needsGuild && guilds?.length === 0 && (
            <div className="text-sm text-muted-foreground">
              Add your bot to a server first, so you can pick its channels and
              roles here.
            </div>
          )}
          {needsGuild && !!guilds && guilds.length > 1 && (
            <div className="space-y-1.5">
              <Label>Server</Label>
              <div className="text-xs text-muted-foreground">
                The server the settings below are picked from.
              </div>
              <GuildSelect
                value={guildId}
                onChange={(id) => {
                  setGuildId(id);
                  // Picks from another server don't exist in this one.
                  setValues((v) => {
                    const next = { ...v };
                    for (const s of template.settings) {
                      if (pickerTypes.includes(s.type)) delete next[s.key];
                    }
                    return next;
                  });
                }}
              />
            </div>
          )}
          {template.settings.map((setting) => (
            <div key={setting.key} className="space-y-1.5">
              <Label htmlFor={`template-setting-${setting.key}`}>
                {setting.label}
                {!setting.required && (
                  <span className="text-muted-foreground font-normal ml-1.5">
                    optional
                  </span>
                )}
              </Label>
              <div className="text-xs text-muted-foreground">
                {setting.description}
              </div>
              <SettingInput
                setting={setting}
                guildId={guildId}
                value={values[setting.key] || ""}
                onChange={(value) =>
                  setValues((v) => ({ ...v, [setting.key]: value }))
                }
              />
            </div>
          ))}
        </div>
      )}

      <div className="text-sm font-medium mb-2.5">Include</div>
      <div className="space-y-2.5 mb-5">
        {template.commands.map((command) => {
          const key = command.key;
          const exists = commandExists(command);
          return (
            <label
              key={key}
              className="flex items-center gap-2.5 text-sm cursor-pointer"
            >
              <Checkbox
                checked={!exists && !skipped.has(key)}
                disabled={exists}
                onCheckedChange={(checked) => toggle(key, !!checked)}
              />
              <span className={exists ? "text-muted-foreground" : ""}>
                /{command.name}
              </span>
              {exists && (
                <span className="text-xs text-muted-foreground ml-auto">
                  already in your app
                </span>
              )}
            </label>
          );
        })}
        {template.eventListeners.map((listener) => {
          const key = listener.key;
          const exists = listenerExists(listener);
          return (
            <label
              key={key}
              className="flex items-center gap-2.5 text-sm cursor-pointer"
            >
              <Checkbox
                checked={!exists && !skipped.has(key)}
                disabled={exists}
                onCheckedChange={(checked) => toggle(key, !!checked)}
              />
              <span className={exists ? "text-muted-foreground" : ""}>
                {discordEventType(listener.type)?.label ?? listener.type}
              </span>
              {exists && (
                <span className="text-xs text-muted-foreground ml-auto">
                  already in your app
                </span>
              )}
            </label>
          );
        })}
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

// Settings that are picked from the server.
const pickerTypes: TemplateSetting["type"][] = ["channel", "role"];

function SettingInput({
  setting,
  guildId,
  value,
  onChange,
}: {
  setting: TemplateSetting;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  const id = `template-setting-${setting.key}`;

  switch (setting.type) {
    case "channel":
      return (
        <ChannelSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
        />
      );
    case "role":
      return (
        <RoleSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
        />
      );
    case "textarea":
      return (
        <Textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    default:
      return (
        <Input
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}
