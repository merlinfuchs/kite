import { useCommandsDeployMutation } from "@/lib/api/mutations";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { App, Command } from "@/lib/types/wire.gen";
import {
  AlertCircleIcon,
  CheckIcon,
  PauseIcon,
  RefreshCwIcon,
} from "lucide-react";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { useTrack } from "@/lib/analytics";

export type CommandDeployState =
  | "pending"
  | "stopped"
  | "error"
  | "deployed"
  | "disabled";

// Pass the command for its own state. Disabled commands aren't in Discord.
export function commandDeployState(
  app: App,
  command?: Command
): CommandDeployState {
  if (command && !command.enabled) {
    return "disabled";
  }
  if (app.commands_deploy_pending) {
    return app.enabled ? "pending" : "stopped";
  }
  return app.commands_deploy_error ? "error" : "deployed";
}

// Whether Discord rejected the commands because of the command with this name.
// Subcommands are deployed as part of their root command.
export function isCommandDeployErrorCause(app: App, commandName: string) {
  return app.commands_deploy_error?.command === commandName.split(" ")[0];
}

// Sends the commands to Discord again, even if they look unchanged.
export function useRedeployCommands(onSuccess?: () => void) {
  const deployMutation = useCommandsDeployMutation(useAppId());
  const track = useTrack();

  function redeploy() {
    track("commands_synced");
    deployMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          onSuccess?.();
        } else {
          toast.error(
            `Failed to sync commands: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return { redeploy, isPending: deployMutation.isPending };
}

export default function CommandDeployStatus() {
  const app = useApp();
  const { redeploy, isPending } = useRedeployCommands();

  if (!app) return null;

  const state = commandDeployState(app);

  if (state === "error") {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon className="h-4 w-4" />
        <AlertTitle>
          Your commands couldn&apos;t be updated in Discord
        </AlertTitle>
        <AlertDescription className="space-y-3">
          <p>{app.commands_deploy_error?.message}</p>
          <p className="text-foreground/80">
            Discord still shows your commands as they were before. Fix the
            problem and save again, or try again if you think it was a temporary
            error.
          </p>
          <LoadingButton
            size="sm"
            variant="outline"
            loading={isPending}
            onClick={redeploy}
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
          {/* An error while still pending is temporary, and retried. */}
          {app.commands_deploy_error?.message ??
            "Updating your commands in Discord..."}
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
