import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import CommandDeployStatus from "./CommandDeployStatus";
import { Command } from "@/lib/types/wire.gen";

export default function CommandList() {
  const commands = useCommands();
  const visibleCommands =
    commands?.filter((command): command is Command => command !== undefined) ??
    [];

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
          <CommandDeployStatus />
          {visibleCommands.length === 0 ? (
            <AppEmptyPlaceholder
              title="No commands yet"
              description="Create a slash command or a context menu command for users and messages."
            />
          ) : (
            visibleCommands.map((command) => (
              <CommandListEntry command={command} key={command.id} />
            ))
          )}

          <div className="flex gap-5 flex-col md:flex-row">
            <CommandCreateDialog>
              <Button>Create command</Button>
            </CommandCreateDialog>
            <FlowImportDialog kind="command">
              <Button variant="outline">Import command</Button>
            </FlowImportDialog>
          </div>
        </>
      )}
    </AutoAnimate>
  );
}
