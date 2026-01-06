import { usePluginInstances, usePlugins } from "@/lib/hooks/api";
import { PluginListEntry } from "./PluginListEntry";
import { useMemo, useState } from "react";
import { CommandDeployDialog } from "./CommandDeployDialog";
import { Button } from "../ui/button";

export function PluginList() {
  const plugins = usePlugins();
  const pluginInstances = usePluginInstances();

  const pluginsWithInstances = useMemo(() => {
    if (!plugins) return [];

    return plugins.map((plugin) => {
      const instance = pluginInstances?.find(
        (instance) => instance!.plugin_id === plugin!.id
      );
      return { ...plugin!, instance: instance };
    });
  }, [plugins, pluginInstances]);

  const hasUndeployedPlugins = pluginsWithInstances.some(
    (plugin) =>
      plugin!.instance &&
      new Date(plugin!.instance.updated_at) >
        new Date(plugin!.instance.last_deployed_at || 0)
  );

  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

  return (
    <div className="space-y-5">
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
        {pluginsWithInstances.map((plugin, i) => (
          <PluginListEntry key={i} plugin={plugin} instance={plugin.instance} />
        ))}
      </div>
      <CommandDeployDialog
        open={deployDialogOpen}
        onOpenChange={setDeployDialogOpen}
      >
        <Button disabled={!hasUndeployedPlugins} variant="destructive">
          Deploy all commands
        </Button>
      </CommandDeployDialog>
    </div>
  );
}
