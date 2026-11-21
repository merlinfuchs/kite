import { useBundleRestoreMutation } from "@/lib/api/mutations";
import { commandLabel } from "@/lib/flow/commandTypes";
import { discordEventType } from "@/lib/flow/eventTypes";
import { restoreSelection } from "@/lib/flow/templateInstall";
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
import { ItemRow } from "./TemplateInstallCard";
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
  const commandNames = useMemo(
    () => new Set(commands?.flatMap((c) => (c ? [c.name] : []))),
    [commands]
  );

  const selection = restoreSelection(template, appIds ?? new Set(), skipped);
  const { installedKeys, parameters } = selection;
  const exceeded = useExceededLimit({
    commands: selection.commands.length,
    eventListenerSources: selection.eventListeners.map((l) => l.source),
    messages: selection.messages.length,
    variables: selection.variables.length,
  });

  const deleted = {
    commands: template.commands.filter((c) => !installedKeys.has(c.key)),
    eventListeners: template.eventListeners.filter(
      (l) => !installedKeys.has(l.key)
    ),
    messages: template.messages.filter((m) => !installedKeys.has(m.key)),
    variables: template.variables.filter((v) => !installedKeys.has(v.key)),
  };
  const items = [
    ...selection.commands,
    ...selection.eventListeners,
    ...selection.messages,
    ...selection.variables,
  ].map((i) => i.key);

  if (!appIds || !Object.values(deleted).some((d) => d.length)) return null;

  function toggle(key: string, included: boolean) {
    const next = new Set(skipped);
    if (included) next.delete(key);
    else next.add(key);
    setSkipped(next);
  }

  async function restore() {
    const res = await restoreMutation.mutateAsync({
      bundle_version: template.version,
      parameters: Object.fromEntries(
        parameters.flatMap((p) =>
          values[p.key] ? [[p.key, values[p.key]]] : []
        )
      ),
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

      <div className="space-y-2.5 mb-5">
        {deleted.commands.map((command) => (
          <ItemRow
            key={command.key}
            exists={false}
            included={!skipped.has(command.key)}
            onToggle={(included) => toggle(command.key, included)}
            note={
              !skipped.has(command.key) && commandNames.has(command.name)
                ? "name taken, gets a number"
                : undefined
            }
          >
            {commandLabel(command.name, command.flowSource)}
          </ItemRow>
        ))}
        {deleted.eventListeners.map((listener) => (
          <ItemRow
            key={listener.key}
            exists={false}
            included={!skipped.has(listener.key)}
            onToggle={(included) => toggle(listener.key, included)}
            description={listener.description}
          >
            {discordEventType(listener.type)?.label ?? listener.type}
          </ItemRow>
        ))}
        {deleted.messages.map((message) => (
          <ItemRow key={message.key} exists={false}>
            {message.name} message
          </ItemRow>
        ))}
        {deleted.variables.map((variable) => (
          <ItemRow key={variable.key} exists={false}>
            {variable.name} stored variable
          </ItemRow>
        ))}
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
