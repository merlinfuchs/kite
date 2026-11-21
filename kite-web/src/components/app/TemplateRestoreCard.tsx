import { useBundleRestoreMutation } from "@/lib/api/mutations";
import { installRequest, restoreSelection } from "@/lib/flow/templateInstall";
import { parameterDefaults, Template } from "@/lib/flow/templates";
import {
  useCommands,
  useEventListeners,
  useMessages,
  useVariables,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { PlanLimitNotice, useExceededLimit } from "./PlanLimit";
import { TemplateItemRows } from "./TemplateInstallCard";
import {
  TemplateParameterFields,
  useTemplateGuild,
} from "./TemplateParameterFields";

// TemplateRestoreCard is for the app a template is made from, to add back the
// items of the template it deleted. They're the originals again, so publishing
// an update still updates them in the apps that added the template. It shows
// nothing while the app has everything.
export function TemplateRestoreCard({ template }: { template: Template }) {
  const appId = useAppId();
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();
  const variables = useVariables();
  const restoreMutation = useBundleRestoreMutation(appId, template.id);

  const [values, setValues] = useState(() =>
    parameterDefaults(template.parameters)
  );
  const [guildId, setGuildId] = useTemplateGuild();
  const [skipped, setSkipped] = useState<Set<string>>(new Set());

  const appIds = useMemo(() => {
    if (!commands || !eventListeners || !messages || !variables) return;
    return new Set(
      [...commands, ...eventListeners, ...messages, ...variables].flatMap((i) =>
        i ? [i.id] : []
      )
    );
  }, [commands, eventListeners, messages, variables]);
  const selection = restoreSelection(template, appIds ?? new Set(), skipped);
  const exceeded = useExceededLimit({
    commands: selection.commands.length,
    eventListenerSources: selection.eventListeners.map((l) => l.source),
    messages: selection.messages.length,
    variables: selection.variables.length,
  });
  // Restoring sends the same items and parameters adding would.
  const { items, parameters } = installRequest(template, selection, values, {});

  const deleted = [
    ...template.commands,
    ...template.eventListeners,
    ...template.messages,
    ...template.variables,
  ].some((i) => !appIds?.has(i.key));
  if (!appIds || !deleted) return null;

  async function restore() {
    const res = await restoreMutation.mutateAsync({
      bundle_version: template.version,
      parameters,
      items,
    });
    if (!res.success) {
      toast.error(
        `Failed to restore the items: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(
      `Restored ${items.length} ${items.length === 1 ? "item" : "items"}`
    );
    setSkipped(new Set());
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Restore deleted items</div>
      <p className="text-sm text-muted-foreground mb-4">
        This app doesn&apos;t have some items of the template anymore. Restore
        them as they are in version {template.version}. They stay part of the
        template, so your next update also updates them in apps that added it.
      </p>

      {!!selection.parameters.length && (
        <div className="mb-5">
          <TemplateParameterFields
            parameters={selection.parameters}
            values={values}
            onChange={setValues}
            guildId={guildId}
            onGuildChange={setGuildId}
          />
        </div>
      )}

      <div className="space-y-2.5 mb-5">
        <TemplateItemRows
          template={template}
          installedKeys={appIds}
          skipped={skipped}
          onSkippedChange={setSkipped}
          hideExisting
        />
      </div>

      {exceeded && (
        <PlanLimitNotice exceeded={exceeded} from="template" className="mb-5" />
      )}

      {items.length === 0 ? (
        <Button className="w-full" disabled>
          Nothing selected
        </Button>
      ) : (
        <LoadingButton
          className="w-full"
          onClick={restore}
          disabled={!!exceeded}
          loading={restoreMutation.isPending}
        >
          Restore {items.length} {items.length === 1 ? "item" : "items"}
        </LoadingButton>
      )}
    </Card>
  );
}
