import { Template } from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import {
  useBundleInstallDeleteMutation,
  useBundleInstallUpdateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useBundleInstalls } from "@/lib/hooks/api";
import { BundleInstall } from "@/lib/types/wire.gen";
import { track } from "@/lib/analytics";
import { ArrowUpCircleIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import ConfirmDialog from "../common/ConfirmDialog";
import LoadingButton from "../common/LoadingButton";
import {
  TemplateSettingFields,
  useTemplateGuild,
} from "./TemplateSettingFields";

// TemplateInstalledCard shows that the app has the template, with updating it
// to a new version and removing it.
export function TemplateInstalledCard({ template }: { template: Template }) {
  const appId = useAppId();
  const installs = (useBundleInstalls() ?? []).filter(
    (i): i is BundleInstall => i?.bundle_id === template.id
  );
  const updateMutation = useBundleInstallUpdateMutation(appId);
  const deleteMutation = useBundleInstallDeleteMutation(appId);

  const [overwrite, setOverwrite] = useState<Set<string>>(new Set());
  const [values, setValues] = useState<Record<string, string>>({});
  const [guildId, setGuildId] = useTemplateGuild();

  const items = installs.flatMap((i) =>
    i.items.filter((item) => !item.deleted)
  );
  if (!items.length) return null;

  const outdated = installs.filter((i) => i.bundle_version < template.version);
  // Items that are no longer in the template stay as they are.
  const updatable = outdated.flatMap((i) =>
    i.items.filter((item) => !item.deleted && templateItem(template, item.key))
  );
  const edited = updatable.filter((item) => item.edited);
  const replaced = updatable.filter(
    (item) => !item.edited || overwrite.has(item.key)
  );

  // Settings the new version uses that an install has no value for.
  const usedKeys = new Set(
    replaced.flatMap((item) => templateItem(template, item.key)!.settingKeys)
  );
  const missingSettings = template.settings.filter(
    (s) => usedKeys.has(s.key) && outdated.some((i) => !i.settings[s.key])
  );

  async function update() {
    for (const install of outdated) {
      const res = await updateMutation.mutateAsync({
        installId: install.id,
        bundle_version: template.version,
        settings: Object.fromEntries(
          missingSettings.flatMap((s) =>
            values[s.key] ? [[s.key, values[s.key]]] : []
          )
        ),
        overwrite: Array.from(overwrite),
      });
      if (!res.success) {
        toast.error(
          `Failed to update ${template.name}: ${res.error.message} (${res.error.code})`
        );
        return;
      }
    }
    track("template_updated", {
      template: template.id,
      template_version: template.version,
    });
    setOverwrite(new Set());
    toast.success(`Updated ${template.name} to version ${template.version}`);
  }

  async function remove() {
    for (const install of installs) {
      const res = await deleteMutation.mutateAsync(install.id);
      if (!res.success) {
        toast.error(
          `Failed to remove ${template.name}: ${res.error.message} (${res.error.code})`
        );
        return;
      }
    }
    track("template_removed", { template: template.id });
    toast.success(`Removed ${template.name} from your app`);
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">In your app</div>
      <p className="text-sm text-muted-foreground mb-4">
        {items.length} {items.length === 1 ? "item" : "items"} of this template
        are in your app.
      </p>

      {!!outdated.length && (
        <div className="mb-5">
          <div className="flex items-center gap-2 text-sm font-medium mb-1.5">
            <ArrowUpCircleIcon className="h-4 w-4 text-primary" />
            Version {template.version} is available
          </div>
          <p className="text-sm text-muted-foreground mb-3">
            Updating replaces what you didn&apos;t change with the new version.
            Messages already posted in Discord stay as they are.
          </p>

          {!!edited.length && (
            <div className="mb-4">
              <div className="text-sm mb-2">
                You changed these, so they&apos;re kept unless you replace them:
              </div>
              <div className="space-y-2.5">
                {edited.map((item) => (
                  <label
                    key={item.id}
                    className="flex items-center gap-2.5 text-sm cursor-pointer"
                  >
                    <Checkbox
                      checked={overwrite.has(item.key)}
                      onCheckedChange={(checked) => {
                        const next = new Set(overwrite);
                        if (checked) next.add(item.key);
                        else next.delete(item.key);
                        setOverwrite(next);
                      }}
                    />
                    <span>{itemLabel(template, item.key)}</span>
                    <span className="text-xs text-muted-foreground ml-auto">
                      replace my changes
                    </span>
                  </label>
                ))}
              </div>
            </div>
          )}

          {!!missingSettings.length && (
            <div className="mb-4">
              <TemplateSettingFields
                settings={missingSettings}
                values={values}
                onChange={setValues}
                guildId={guildId}
                onGuildChange={setGuildId}
              />
            </div>
          )}

          <LoadingButton
            className="w-full"
            onClick={update}
            loading={updateMutation.isPending}
          >
            Update to version {template.version}
          </LoadingButton>
        </div>
      )}

      <ConfirmDialog
        title={`Remove ${template.name}?`}
        description={`This deletes ${new Intl.ListFormat("en").format(
          items.map((item) => itemLabel(template, item.key))
        )} from your app, including your changes to them.`}
        onConfirm={remove}
      >
        <Button
          variant="outline"
          className="w-full"
          disabled={deleteMutation.isPending}
        >
          Remove from app
        </Button>
      </ConfirmDialog>
    </Card>
  );
}

function templateItem(template: Template, key: string) {
  const command = template.commands.find((c) => c.key === key);
  if (command) return { label: `/${command.name}`, ...command };
  const listener = template.eventListeners.find((l) => l.key === key);
  if (listener) {
    return {
      label: discordEventType(listener.type)?.label ?? listener.type,
      ...listener,
    };
  }
  const message = template.messages.find((m) => m.key === key);
  if (message) {
    return {
      label: `${message.name} message`,
      ...message,
      settingKeys: message.sendTo
        ? [...message.settingKeys, message.sendTo]
        : message.settingKeys,
    };
  }
}

function itemLabel(template: Template, key: string) {
  return templateItem(template, key)?.label ?? key;
}
