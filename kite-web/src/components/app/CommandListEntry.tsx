import {
  useCommandDeleteMutation,
  useCommandUpdateEnabledMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { Command } from "@/lib/types/wire.gen";
import { formatDateTime } from "@/lib/utils";
import {
  CheckIcon,
  CircleDotIcon,
  CopyPlusIcon,
  EllipsisIcon,
  SlashSquareIcon,
  Trash2Icon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useCallback, useMemo } from "react";
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
import CommandDuplicateDialog from "./CommandDuplicateDialog";

export default function CommandListEntry({ command }: { command: Command }) {
  const router = useRouter();

  const appId = useAppId();

  const deleteMutation = useCommandDeleteMutation(appId, command.id);
  const updateEnabledMutation = useCommandUpdateEnabledMutation(
    appId,
    command.id
  );

  const remove = useCallback(() => {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Command deleted!");
        } else {
          toast.error(
            `Failed to delete command: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }, [deleteMutation]);

  const toggleEnabled = useCallback(() => {
    updateEnabledMutation.mutate({
      enabled: !command.enabled,
    });
  }, [updateEnabledMutation, command.enabled]);

  const changesDeployed = useMemo(
    () =>
      new Date(command.updated_at) <= new Date(command.last_deployed_at || 0),
    [command]
  );

  return (
    <Card className="relative">
      <div className="absolute top-0 right-0 py-3 pr-3 h-full flex flex-col justify-between">
        <div className="flex items-center space-x-2">
          {changesDeployed ? (
            <Tooltip>
              <TooltipTrigger>
                <CheckIcon className="h-5 w-5 text-green-500" />
              </TooltipTrigger>
              <TooltipContent>
                <div className="text-foreground/90">
                  All changes have been deployed!
                </div>
              </TooltipContent>
            </Tooltip>
          ) : (
            <Tooltip>
              <TooltipTrigger>
                <CircleDotIcon className="h-5 w-5 text-orange-500" />
              </TooltipTrigger>
              <TooltipContent>
                <div className="text-foreground/90">
                  Most recent changes will be deployed soon.
                </div>
              </TooltipContent>
            </Tooltip>
          )}
          <div className="text-sm text-muted-foreground">
            {formatDateTime(new Date(command.updated_at))}
          </div>
        </div>
        <div className="flex justify-end">
          <Switch checked={command.enabled} onCheckedChange={toggleEnabled} />
        </div>
      </div>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <SlashSquareIcon className="h-5 w-5 text-muted-foreground" />
          <div>{command.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          {command.description}
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
                  Delete Command
                </DropdownMenuItem>
              </ConfirmDialog>
              <CommandDuplicateDialog command={command}>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <CopyPlusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Duplicate Command
                </DropdownMenuItem>
              </CommandDuplicateDialog>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardFooter>
    </Card>
  );
}
