import {
  prepareTemplateFlow,
  Template,
  templateItemKey,
} from "@/lib/flow/templates";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import { useBundleInstallCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useCommands } from "@/lib/hooks/api";
import { track } from "@/lib/analytics";
import { CircleCheckIcon } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";

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

  const bundleInstallMutation = useBundleInstallCreateMutation(appId);

  const [inputValues, setInputValues] = useState<Record<string, string>>({});
  const [skipped, setSkipped] = useState<Set<string>>(new Set());
  const [installed, setInstalled] = useState<{
    commands: number;
    eventListeners: number;
  } | null>(null);

  const commands = template.commands.filter(
    (c) => !skipped.has(`c:${c.name}`) && !existingCommands.has(c.name)
  );
  const eventListeners = template.eventListeners.filter(
    (_, i) => !skipped.has(`e:${i}`)
  );
  const itemCount = commands.length + eventListeners.length;

  function toggle(key: string, included: boolean) {
    const next = new Set(skipped);
    if (included) next.delete(key);
    else next.add(key);
    setSkipped(next);
  }

  async function install() {
    for (const input of template.inputs) {
      if (input.required && !inputValues[input.key]?.trim()) {
        toast.error(`${input.label} is required`);
        return;
      }
    }

    const res = await bundleInstallMutation.mutateAsync({
      bundle_id: template.id,
      bundle_version: template.version,
      name: template.name,
      description: template.description,
      settings: Object.fromEntries(
        template.inputs.flatMap((input) =>
          inputValues[input.key] ? [[input.key, inputValues[input.key]]] : []
        )
      ),
      commands: commands.map((c) => ({
        key: templateItemKey(c),
        command: {
          enabled: true,
          flow_source: prepareTemplateFlow(c.flowSource(inputValues)),
        },
      })),
      event_listeners: eventListeners.map((l) => ({
        key: templateItemKey(l),
        event_listener: {
          enabled: true,
          source: l.source,
          flow_source: prepareTemplateFlow(l.flowSource(inputValues)),
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

      {!!template.inputs.length && (
        <div className="space-y-4 mb-5">
          {template.inputs.map((input) => {
            const id = `template-input-${input.key}`;
            const props = {
              id,
              value: inputValues[input.key] || "",
              onChange: (
                e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
              ) =>
                setInputValues({ ...inputValues, [input.key]: e.target.value }),
            };

            return (
              <div key={input.key} className="space-y-1.5">
                <Label htmlFor={id}>
                  {input.label}
                  {!input.required && (
                    <span className="text-muted-foreground font-normal ml-1.5">
                      optional
                    </span>
                  )}
                </Label>
                <div className="text-xs text-muted-foreground">
                  {input.description}
                </div>
                {input.type === "textarea" ? (
                  <Textarea {...props} />
                ) : (
                  <Input {...props} />
                )}
              </div>
            );
          })}
        </div>
      )}

      <div className="text-sm font-medium mb-2.5">Include</div>
      <div className="space-y-2.5 mb-5">
        {template.commands.map((command) => {
          const key = `c:${command.name}`;
          const exists = existingCommands.has(command.name);
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
        {template.eventListeners.map((listener, i) => {
          const key = `e:${i}`;
          return (
            <label
              key={key}
              className="flex items-center gap-2.5 text-sm cursor-pointer"
            >
              <Checkbox
                checked={!skipped.has(key)}
                onCheckedChange={(checked) => toggle(key, !!checked)}
              />
              {discordEventType(listener.type)?.label ?? listener.type}
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
