import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder, {
  AppEmptyTemplatesLink,
} from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import CommandDeployStatus from "./CommandDeployStatus";
import { Command } from "@/lib/types/wire.gen";
import { SquareSlashIcon } from "lucide-react";
import SearchableList, { ListActions } from "./SearchableList";
import { commandType, isSlashCommand } from "@/lib/flow/commandType";

export default function CommandList() {
  const commands = useCommands();
  const visibleCommands =
    commands?.filter((command): command is Command => command !== undefined) ??
    [];

  const commandActions = (
    <>
      <CommandCreateDialog>
        <Button>Create command</Button>
      </CommandCreateDialog>
      <FlowImportDialog kind="command">
        <Button variant="outline">Import command</Button>
      </FlowImportDialog>
    </>
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
          {visibleCommands.length === 0 ? (
            <AppEmptyPlaceholder
              icon={SquareSlashIcon}
              title="No commands yet"
              description="Create a slash command or a context menu command for users and messages, then build what it does in the flow editor."
              action={<ListActions>{commandActions}</ListActions>}
              footer={<AppEmptyTemplatesLink />}
            />
          ) : (
            <SearchableList
              items={visibleCommands}
              // "mod ban" becomes "/mod /ban", so "/ban" finds the subcommand.
              fields={(command) => {
                const isSlash = isSlashCommand(
                  commandType(command.flow_source)
                );
                return [
                  (isSlash ? "/" : "") +
                    (isSlash
                      ? command.name.split(" ").join(" /")
                      : command.name),
                  command.description,
                ];
              }}
              renderItem={(command) => (
                <CommandListEntry command={command} key={command.id} />
              )}
              placeholder="Search commands"
              actions={commandActions}
            />
          )}
        </>
      )}
    </AutoAnimate>
  );
}
