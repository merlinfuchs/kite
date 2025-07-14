import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import DynamicIcon from "../icons/DynamicIcon";
import { Plugin, PluginInstance } from "@/lib/types/wire.gen";
import { Button } from "../ui/button";
import { Switch } from "../ui/switch";
import { usePluginInstanceUpdateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { PluginConfigureDialog } from "./PluginConfigureDialog";

export function PluginListEntry({
  plugin,
  instance,
}: {
  plugin: Plugin;
  instance?: PluginInstance;
}) {
  const updateMutation = usePluginInstanceUpdateMutation(useAppId(), plugin.id);

  function toggleEnabled(enabled: boolean) {
    if (!instance) return;

    updateMutation.mutate({
      enabled: enabled,
      config: instance.config,
      enabled_resource_ids: instance.enabled_resource_ids,
    });
  }

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row gap-4 p-4 flex-auto">
        <div className="h-10 w-10 bg-primary/40 flex-none rounded-md flex items-center justify-center">
          <DynamicIcon
            name={plugin.metadata.icon as any}
            className="w-6 h-6 text-primary"
          />
        </div>
        <div className="flex-auto">
          <div className="flex items-start justify-between">
            <CardTitle className="mb-2">{plugin.metadata.name}</CardTitle>
            <Switch
              disabled={!instance}
              checked={!!instance?.enabled}
              onCheckedChange={toggleEnabled}
            />
          </div>
          <CardDescription>{plugin.metadata.description}</CardDescription>
        </div>
      </CardHeader>
      <CardFooter className="p-4 pt-1 flex items-center justify-end">
        <PluginConfigureDialog plugin={plugin} instance={instance}>
          <Button variant="outline">Configure</Button>
        </PluginConfigureDialog>
      </CardFooter>
    </Card>
  );
}
