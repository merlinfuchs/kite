import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import CommandDeployStatus from "./CommandDeployStatus";

export default function CommandList() {
  const commands = useCommands();

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
          {commands.length === 0 ? (
            <AppEmptyPlaceholder
              title="There are no commands"
              description="You can start now by creating the first command!"
            />
          ) : (
            commands.map((command, i) => (
              <CommandListEntry command={command!} key={i} />
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
