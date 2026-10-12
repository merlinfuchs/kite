import { useCommandsDeployMutation } from "@/lib/api/mutations";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { App } from "@/lib/types/wire.gen";
import {
  AlertCircleIcon,
  CheckIcon,
  PauseIcon,
  RefreshCwIcon,
} from "lucide-react";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";

export type CommandDeployState = "pending" | "stopped" | "error" | "deployed";

// Kite deploys the commands to Discord in the background after every change.
export function commandDeployState(app: App): CommandDeployState {
  if (app.commands_deploy_pending) {
    return app.enabled ? "pending" : "stopped";
  }
  return app.commands_deploy_error ? "error" : "deployed";
}

// Whether Discord rejected the commands because of the command with this name.
// Subcommands are deployed as part of their root command.
export function isCommandDeployErrorCause(app: App, commandName: string) {
  return (
    !!app.commands_deploy_error_command &&
    app.commands_deploy_error_command === commandName.split(" ")[0]
  );
}

export default function CommandDeployStatus() {
  const app = useApp();
  const deployMutation = useCommandsDeployMutation(useAppId());

  if (!app) return null;

  function retry() {
    deployMutation.mutate(undefined, {
      onSuccess(res) {
        if (!res.success) {
          toast.error(
            `Failed to deploy commands: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const state = commandDeployState(app);

  if (state === "error") {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon className="h-4 w-4" />
        <AlertTitle>Discord rejected your latest changes</AlertTitle>
        <AlertDescription className="space-y-3">
          <p>{app.commands_deploy_error}</p>
          <p className="text-foreground/80">
            Discord still shows your commands as they were before. Fix the
            problem and save again, or try again if you think it was a temporary
            error.
          </p>
          <LoadingButton
            size="sm"
            variant="outline"
            loading={deployMutation.isPending}
            onClick={retry}
          >
            Try again
          </LoadingButton>
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="flex items-center gap-2 text-sm text-muted-foreground">
      {state === "pending" ? (
        <>
          <RefreshCwIcon className="h-4 w-4 flex-none animate-spin" />
          Updating your commands in Discord...
        </>
      ) : state === "stopped" ? (
        <>
          <PauseIcon className="h-4 w-4 flex-none" />
          Your app is stopped. Your command changes go to Discord once it runs
          again.
        </>
      ) : (
        <>
          <CheckIcon className="h-4 w-4 flex-none text-green-500" />
          Your commands are up to date in Discord. Restart Discord if a change
          doesn&apos;t show up.
        </>
      )}
    </div>
  );
}
