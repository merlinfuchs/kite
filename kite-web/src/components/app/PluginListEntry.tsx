import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Plugin } from "@/lib/types/wire.gen";
import { BlocksIcon, CalculatorIcon } from "lucide-react";
import { PluginConfigureDialog } from "./PluginConfigureDialog";
import { Switch } from "../ui/switch";
import { usePluginInstanceUpdateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { usePluginInstance } from "@/lib/hooks/api";
import { useCallback, useMemo } from "react";
import DynamicIcon from "../icons/DynamicIcon";

export function PluginListEntry({ plugin }: { plugin: Plugin }) {
  const appId = useAppId();

  const instance = usePluginInstance(plugin.id);

  const updateMutation = usePluginInstanceUpdateMutation(appId, plugin.id);

  const enabled = useMemo(() => instance?.enabled ?? false, [instance]);

  const toggleEnabled = useCallback(() => {
    updateMutation.mutate({
      enabled: !enabled,
      config: instance?.config ?? {},
    });
  }, [instance, enabled, updateMutation]);

  return (
    <Card>
      <CardHeader className="flex flex-row gap-4 p-4 items-start">
        <div className="h-10 w-10 bg-primary/40 flex-none rounded-md flex items-center justify-center">
          <CalculatorIcon
            name={plugin.icon as any}
            className="w-6 h-6 text-primary"
          />
        </div>
        <div className="flex-auto">
          <CardTitle className="mb-2">{plugin.name}</CardTitle>
          <CardDescription>{plugin.description}</CardDescription>
        </div>
        <div className="flex-none">
          <Switch checked={enabled} onCheckedChange={toggleEnabled} />
        </div>
      </CardHeader>
      <CardFooter className="p-4 pt-1 flex justify-end">
        <PluginConfigureDialog plugin={plugin}>
          <Button variant="outline">Configure</Button>
        </PluginConfigureDialog>
      </CardFooter>
    </Card>
  );
}
