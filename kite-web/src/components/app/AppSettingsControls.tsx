import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Button } from "../ui/button";
import { useApp } from "@/lib/hooks/api";
import {
  useAppRestartMutation,
  useAppUpdateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";

export default function AppSettingsControls() {
  const app = useApp();

  const appId = useAppId();
  const updateMutation = useAppUpdateMutation(appId);
  const restartMutation = useAppRestartMutation(appId);

  function toggleEnabled() {
    if (!app) return;

    updateMutation.mutate(
      {
        name: app.name,
        description: app.description,
        enabled: !app.enabled,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success(app.enabled ? "App disabled!" : "App enabled!");
          } else {
            toast.error(
              `Failed to update app: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  function restart() {
    restartMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("App is restarting, it will be back in a few seconds!");
        } else {
          toast.error(
            `Failed to restart app: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <CardContent className="pt-6 space-y-5">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Start App</div>
            <div className="text-muted-foreground">
              Starting your app will make it appear as online in Discord and
              allow users to interact with it.
            </div>
          </div>
          <Button
            disabled={app?.enabled}
            variant="secondary"
            onClick={toggleEnabled}
          >
            Start App
          </Button>
        </div>
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Stop App</div>
            <div className="text-muted-foreground">
              Stopping your app will make it appear as offline in Discord and
              prevent users from interacting with it.
            </div>
          </div>
          <Button
            disabled={!app?.enabled}
            variant="secondary"
            onClick={toggleEnabled}
          >
            Stop App
          </Button>
        </div>
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Restart App</div>
            <div className="text-muted-foreground">
              Restarting your app reconnects it to Discord. It will appear as
              offline for a few seconds.
            </div>
          </div>
          <Button
            disabled={!app?.enabled || restartMutation.isPending}
            variant="secondary"
            onClick={restart}
          >
            Restart App
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
