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
        {/* Never disabled: removing a plugin instance deletes the row that
            "has undeployed changes" was derived from, so gating on it made it
            impossible to remove a disabled plugin's commands from Discord. */}
        <Button variant="destructive">Deploy all commands</Button>
      </CommandDeployDialog>
    </div>
  );
}
