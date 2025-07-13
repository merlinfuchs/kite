import { usePluginInstances, usePlugins } from "@/lib/hooks/api";
import { PluginListEntry } from "./PluginListEntry";
import { useMemo } from "react";

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

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      {pluginsWithInstances.map((plugin, i) => (
        <PluginListEntry key={i} plugin={plugin} instance={plugin.instance} />
      ))}
    </div>
  );
}
