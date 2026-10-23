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
import { SquareSlashIcon } from "lucide-react";
import SearchableList, { ListActions } from "./SearchableList";

export default function CommandList() {
  const commands = useCommands();

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
          {commands.length === 0 ? (
            <AppEmptyPlaceholder
              icon={SquareSlashIcon}
              title="No commands yet"
              description="Commands are what your users type in Discord, like /ban or /help. Create one and build what it does in the flow editor."
              action={<ListActions>{commandActions}</ListActions>}
              footer={<AppEmptyTemplatesLink />}
            />
          ) : (
            <SearchableList
              items={commands}
              // "mod ban" becomes "/mod /ban", so "/ban" finds the subcommand.
              fields={(command) => [
                "/" + command.name.split(" ").join(" /"),
                command.description,
              ]}
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
