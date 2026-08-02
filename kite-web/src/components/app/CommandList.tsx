import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import { useCommands } from "@/lib/hooks/api";
import { CommandDeployDialog } from "./CommandDeployDialog";
import { useState } from "react";

export default function CommandList() {
  const commands = useCommands();

  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!commands ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : (
        <>
          {commands.length === 0 ? (
            <AppEmptyPlaceholder
              title="There are no commands"
              description="You can start now by creating the first command! If you deleted commands that still show up in Discord, deploy to remove them."
            />
          ) : (
            commands.map((command, i) => (
              <CommandListEntry command={command!} key={i} />
            ))
          )}

          {/* The deploy button is never disabled: deleting a command doesn't
              change any remaining command's updated_at, and deleting the last
              one leaves no command to compare at all. Gating on "has
              undeployed changes" made deleted commands unremovable. */}
          <div className="flex gap-5 justify-between flex-col md:flex-row">
            <CommandCreateDialog>
              <Button>Create command</Button>
            </CommandCreateDialog>

            <CommandDeployDialog
              open={deployDialogOpen}
              onOpenChange={setDeployDialogOpen}
            >
              <Button variant="destructive">Deploy all commands</Button>
            </CommandDeployDialog>
          </div>
        </>
      )}
    </AutoAnimate>
  );
}
