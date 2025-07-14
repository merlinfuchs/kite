import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
  usePluginInstanceCreateMutation,
  usePluginInstanceUpdateMutation,
} from "@/lib/api/mutations";
import { prepareTemplateFlow, Template } from "@/lib/flow/templates";
import { useAppId } from "@/lib/hooks/params";
import { SatelliteDishIcon, SlashSquareIcon } from "lucide-react";
import { ReactNode, useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card, CardDescription, CardTitle } from "../ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Switch } from "../ui/switch";
import { Plugin, PluginInstance } from "@/lib/types/wire.gen";

export function PluginConfigureDialog({
  children,
  plugin,
  instance,
}: {
  children: ReactNode;
  plugin: Plugin;
  instance?: PluginInstance;
}) {
  const appId = useAppId();

  const createMutation = usePluginInstanceCreateMutation(appId);
  const updateMutation = usePluginInstanceUpdateMutation(appId, plugin.id);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [config, setConfig] = useState<Record<string, any>>({});
  const [enabledResourceIds, setEnabledResourceIds] = useState<string[]>([]);

  useEffect(() => {
    if (instance) {
      setEnabled(instance.enabled);
      setConfig(instance.config);
      setEnabledResourceIds(instance.enabled_resource_ids);
    } else {
      setEnabledResourceIds([
        ...plugin.commands.map((c) => c.id),
        ...plugin.events.map((e) => e.id),
      ]);
      setEnabled(true);
    }
  }, [plugin, instance]);

  const handleSave = useCallback(() => {
    if (instance) {
      updateMutation.mutate({
        enabled: enabled,
        config: config,
        enabled_resource_ids: enabledResourceIds,
      });
    } else {
      createMutation.mutate({
        plugin_id: plugin.id,
        config: config,
        enabled_resource_ids: enabledResourceIds,
        enabled: enabled,
      });
    }

    setDialogOpen(false);
  }, [plugin, instance, enabled, config, enabledResourceIds]);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">
            {plugin.metadata.name} Plugin
          </DialogTitle>
          <DialogDescription>{plugin.metadata.description}</DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2 space-y-5">
          <div>
            <div className="font-medium mb-2">Enabled</div>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>

          {!!plugin.config.sections && (
            <div className="flex flex-col gap-3">
              {plugin.config.sections.map((section) => (
                <div key={section.name}>
                  <div className="flex flex-col gap-3">
                    {section.fields.map((field) => (
                      <div key={field.key}>
                        <div className="mb-3">
                          <div className="flex items-center gap-2">
                            <div className="font-medium mb-0.5">
                              {field.name}
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {field.required ? (
                                <span className="text-red-500">required</span>
                              ) : (
                                <span className="text-muted-foreground">
                                  optional
                                </span>
                              )}
                            </div>
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {field.description}
                          </div>
                        </div>
                        <Input
                          value={config[field.key] || ""}
                          onChange={(e) =>
                            setConfig({
                              ...config,
                              [field.key]: e.target.value || undefined,
                            })
                          }
                        />
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}

          {!!plugin.commands.length && (
            <div>
              <div className="mb-3">
                <div className="font-medium mb-0.5">Commands</div>
                <div className="text-sm text-muted-foreground">
                  Select the commands you want to import.
                </div>
              </div>
              <div className="flex flex-col gap-3">
                {plugin.commands.map((command, i) => (
                  <Card className="px-4 pb-4 pt-3" key={i}>
                    <div className="float float-right">
                      <Switch
                        checked={enabledResourceIds.includes(command.id)}
                        onCheckedChange={(checked) => {
                          setEnabledResourceIds(
                            !checked
                              ? enabledResourceIds.filter(
                                  (c) => c !== command.id
                                )
                              : [...enabledResourceIds, command.id]
                          );
                        }}
                      />
                    </div>

                    <div className="flex items-center gap-1.5 mb-0.5">
                      <SlashSquareIcon className="text-muted-foreground h-5 w-5" />
                      <CardTitle className="text-lg">
                        {command.data.name}
                      </CardTitle>
                    </div>
                    <CardDescription className="text-sm">
                      {command.data.description}
                    </CardDescription>
                  </Card>
                ))}
              </div>
            </div>
          )}

          {!!plugin.events.length && (
            <div>
              <div className="mb-3">
                <div className="font-medium mb-0.5">Event Listeners</div>
                <div className="text-sm text-muted-foreground">
                  Select the event listeners you want to import.
                </div>
              </div>
              <div className="flex flex-col gap-3">
                {plugin.events.map((event, i) => (
                  <Card className="px-4 pb-4 pt-3" key={i}>
                    <div className="float float-right">
                      <Switch
                        checked={enabledResourceIds.includes(event.id)}
                        onCheckedChange={(checked) => {
                          setEnabledResourceIds(
                            !checked
                              ? enabledResourceIds.filter((c) => c !== event.id)
                              : [...enabledResourceIds, event.id]
                          );
                        }}
                      />
                    </div>

                    <div className="flex items-center gap-1.5 mb-0.5">
                      <SatelliteDishIcon className="text-muted-foreground h-5 w-5" />
                      <CardTitle className="text-lg">{event.type}</CardTitle>
                    </div>
                    <CardDescription className="text-sm">
                      {event.description}
                    </CardDescription>
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
