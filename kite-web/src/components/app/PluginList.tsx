import { usePlugins } from "@/lib/hooks/api";
import { PluginListEntry } from "./PluginListEntry";

export function PluginList() {
  const plugins = usePlugins();

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      {plugins?.map((plugin, i) => (
        <PluginListEntry key={i} plugin={plugin} />
      ))}
    </div>
  );
}
