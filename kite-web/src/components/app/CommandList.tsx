import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import { receiveCommandShare } from "./CommandShareDialog";
import { useCommands } from "@/lib/hooks/api";
import { CommandDeployDialog } from "./CommandDeployDialog";
import { useState } from "react";

export default function CommandList() {
  const commands = useCommands();

  const cmdCreateButton = (
    <CommandCreateDialog>
      <Button>Create command</Button>
    </CommandCreateDialog>
  );
  
  const cmdImportButton = (
    <CommandShareDialog>
      <Button>Import command</Button>
    </CommandShareDialog>
  );

  const hasUndeployedCommands = commands?.some(
    (command) =>
      new Date(command!.updated_at) > new Date(command!.last_deployed_at || 0)
  );

  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!commands ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : commands.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no commands"
          description="You can start now by creating the first command!"
          action={cmdCreateButton}
        />
      ) : (
        <>
          {commands.map((command, i) => (
            <CommandListEntry command={command!} key={i} />
          ))}

          <div className="flex gap-5 justify-between flex-col md:flex-row">
            {cmdCreateButton}
            {cmdImportButton}

            <CommandDeployDialog
              open={deployDialogOpen}
              onOpenChange={setDeployDialogOpen}
            >
              <Button disabled={!hasUndeployedCommands} variant="destructive">
                Deploy all commands
              </Button>
            </CommandDeployDialog>
          </div>
        </>
      )}
    </AutoAnimate>
  );
}
