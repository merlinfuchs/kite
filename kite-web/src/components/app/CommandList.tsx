import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import CommandDeployStatus from "./CommandDeployStatus";
import { commandType, isSlashCommand } from "@/lib/flow/commandType";

export default function CommandList({
  kind = "command",
}: {
  kind?: "command" | "context_menu";
}) {
  const commands = useCommands()?.filter(
    (c) =>
      c && isSlashCommand(commandType(c.flow_source)) === (kind === "command")
  );

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
              title={
                kind === "context_menu"
                  ? "There are no context menu commands"
                  : "There are no commands"
              }
              description="You can start now by creating the first command!"
            />
          ) : (
            commands.map((command, i) => (
              <CommandListEntry command={command!} key={i} />
            ))
          )}

          <div className="flex gap-5 flex-col md:flex-row">
            <CommandCreateDialog contextMenu={kind === "context_menu"}>
              <Button>
                {kind === "context_menu"
                  ? "Create context menu"
                  : "Create command"}
              </Button>
            </CommandCreateDialog>
            <FlowImportDialog kind="command">
              <Button variant="outline">
                {kind === "context_menu"
                  ? "Import context menu"
                  : "Import command"}
              </Button>
            </FlowImportDialog>
          </div>
        </>
      )}
    </AutoAnimate>
  );
}
