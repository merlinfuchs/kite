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
import { PlanLimitCount, PlanLimitGate, usePlanLimit } from "./PlanLimit";

export default function CommandList() {
  const commands = useCommands();
  const limit = usePlanLimit("max_commands", commands?.length);

  const commandActions = (
    <PlanLimitGate limit={limit} noun="commands" from="commands">
      <CommandCreateDialog>
        <Button>Create command</Button>
      </CommandCreateDialog>
      <FlowImportDialog kind="command">
        <Button variant="outline">Import command</Button>
      </FlowImportDialog>
    </PlanLimitGate>
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
              status={
                <PlanLimitCount limit={limit} noun="commands" from="commands" />
              }
            />
          )}
        </>
      )}
    </AutoAnimate>
  );
}
