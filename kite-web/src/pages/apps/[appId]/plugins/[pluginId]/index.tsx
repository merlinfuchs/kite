import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";
import { usePlugin, usePluginInstance } from "@/lib/hooks/api";
import { useMemo } from "react";

export default function AppPluginPage() {
  const plugin = usePlugin();
  const pluginInstance = usePluginInstance();

  const pluginName = useMemo(() => {
    return plugin?.metadata.name ?? "Unknown Plugin";
  }, [plugin]);

  const breadcrumbs = useMemo(() => {
    return [
      {
        label: pluginName,
      },
    ];
  }, [pluginName]);

  return (
    <AppLayout title={pluginName} breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">{pluginName}</h1>
        <p className="text-muted-foreground text-sm">
          {plugin?.metadata.description}
        </p>
      </div>
      <Separator className="my-8" />
    </AppLayout>
  );
}
