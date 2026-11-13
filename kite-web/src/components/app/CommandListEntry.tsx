import { withoutParameterValues } from "@/lib/flow/parameters";
import {
  useCommandDeleteMutation,
  useCommandUpdateEnabledMutation,
} from "@/lib/api/mutations";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { App, Command } from "@/lib/types/wire.gen";
import { formatDateTime } from "@/lib/utils";
import { commandTypeOf, commandTypeOptions } from "@/lib/flow/commandTypes";
import {
  AlertCircleIcon,
  CheckIcon,
  CircleDotIcon,
  CircleOffIcon,
  RefreshCwIcon,
  CopyPlusIcon,
  EllipsisIcon,
  MessageSquareIcon,
  SlashSquareIcon,
  UserIcon,
  Trash2Icon,
  Share2Icon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useCallback } from "react";
import { toast } from "sonner";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Switch } from "../ui/switch";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import {
  commandDeployState,
  isCommandDeployErrorCause,
} from "./CommandDeployStatus";
import CommandDuplicateDialog from "./CommandDuplicateDialog";
import { PlanLimit, PlanLimitGate } from "./PlanLimit";
import BundleInstallBadge from "./BundleInstallBadge";
import FlowExportDialog from "./FlowExportDialog";

const commandTypeIcons = {
  chat_input: SlashSquareIcon,
  user: UserIcon,
  message: MessageSquareIcon,
};

export default function CommandListEntry({
  command,
  limit,
}: {
  command: Command;
  limit?: PlanLimit;
}) {
  const router = useRouter();

  const appId = useAppId();
  const app = useApp();

  const commandType = commandTypeOf(command.flow_source);
  const TypeIcon = commandTypeIcons[commandType];

  const deleteMutation = useCommandDeleteMutation(appId, command.id);
  const updateEnabledMutation = useCommandUpdateEnabledMutation(
    appId,
    command.id
  );

  const remove = useCallback(() => {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Command deleted.");
        } else {
          toast.error(
            `Failed to delete command: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }, [deleteMutation]);

  const toggleEnabled = useCallback(() => {
    updateEnabledMutation.mutate(
      { enabled: !command.enabled },
      {
        onSuccess(res) {
          if (!res.success) {
            toast.error(
              `Failed to update command: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }, [updateEnabledMutation, command.enabled]);

  return (
    <Card className="relative">
      <div className="flex items-center justify-between px-6 pt-4 sm:absolute sm:top-0 sm:right-0 sm:h-full sm:flex-col sm:items-end sm:p-3">
        <div className="flex items-center space-x-2">
          {app && <DeployIcon app={app} command={command} />}
          <div className="text-sm text-muted-foreground">
            {formatDateTime(new Date(command.updated_at))}
          </div>
        </div>
        <div className="flex justify-end">
          {/* Enabling one more would go over the limit. */}
          <PlanLimitGate
            limit={command.enabled ? undefined : limit}
            noun="commands"
            from="commands"
            countsEnabled
          >
            <Switch checked={command.enabled} onCheckedChange={toggleEnabled} />
          </PlanLimitGate>
        </div>
      </div>
      <CardHeader className="pt-3 sm:pt-6 sm:pr-56">
        <CardTitle className="text-base flex flex-wrap items-center gap-x-2 gap-y-1">
          <div className="flex items-center gap-2 min-w-0">
            <TypeIcon className="h-5 w-5 flex-none text-muted-foreground" />
            <div className="min-w-0 break-words">{command.name}</div>
          </div>
          <BundleInstallBadge bundleInstallId={command.bundle_install_id} />
        </CardTitle>
        <CardDescription className="text-sm">
          {commandType === "chat_input"
            ? command.description
            : commandTypeOptions.find((o) => o.value === commandType)?.label}
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/commands/[cmdId]",
              query: {
                appId: router.query.appId,
                cmdId: command.id,
              },
            }}
          >
            Manage
          </Link>
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="icon" variant="ghost">
              <EllipsisIcon className="h-5 w-5 text-muted-foreground" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuGroup>
              <ConfirmDialog
                title="Are you sure that you want to delete this command?"
                description="This will remove the command from your app and cannot be undone."
                onConfirm={remove}
              >
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Trash2Icon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Delete command
                </DropdownMenuItem>
              </ConfirmDialog>
              <CommandDuplicateDialog command={command}>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <CopyPlusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Duplicate command
                </DropdownMenuItem>
              </CommandDuplicateDialog>
              <FlowExportDialog
                title="Export command"
                type="command"
                shareData={{
                  flow_source: withoutParameterValues(command.flow_source),
                }}
                flows={[command.flow_source]}
              >
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Share2Icon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Export command
                </DropdownMenuItem>
              </FlowExportDialog>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardFooter>
    </Card>
  );
}

function DeployIcon({ app, command }: { app: App; command: Command }) {
  const state = commandDeployState(app, command);

  let icon = <CheckIcon className="h-5 w-5 text-green-500" />;
  let text = "Up to date in Discord.";
  if (state === "pending") {
    icon = (
      <RefreshCwIcon className="h-5 w-5 text-muted-foreground animate-spin" />
    );
    text = "Updating in Discord...";
  } else if (state === "disabled") {
    icon = <CircleOffIcon className="h-5 w-5 text-muted-foreground" />;
    text = "Not in Discord while disabled.";
  } else if (state === "stopped") {
    icon = <CircleDotIcon className="h-5 w-5 text-orange-500" />;
    text = "Goes to Discord once your app runs again.";
  } else if (state === "error") {
    if (isCommandDeployErrorCause(app, command)) {
      icon = <AlertCircleIcon className="h-5 w-5 text-destructive" />;
      text = "Discord rejected this command. See the error above.";
    } else {
      icon = <CircleDotIcon className="h-5 w-5 text-orange-500" />;
      text = "Not updated in Discord until the error above is fixed.";
    }
  }

  return (
    <Tooltip>
      <TooltipTrigger>{icon}</TooltipTrigger>
      <TooltipContent>
        <div className="text-foreground/90">{text}</div>
      </TooltipContent>
    </Tooltip>
  );
}
