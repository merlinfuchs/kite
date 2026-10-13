import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import { CommandDeployDialog } from "./CommandDeployDialog";
import { useState } from "react";
import ListSearchInput, { ListSearchEmpty } from "./ListSearchInput";
import { matchesSearch } from "@/lib/search";

export default function CommandList() {
  const commands = useCommands();

  const [deployDialogOpen, setDeployDialogOpen] = useState(false);
  const [search, setSearch] = useState("");

  const filtered = commands?.filter((command) =>
    matchesSearch(search, [command!.name, command!.description])
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
          {commands.length === 0 ? (
            <AppEmptyPlaceholder
              title="There are no commands"
              description="You can start now by creating the first command! If you deleted commands that still show up in Discord, deploy to remove them."
            />
          ) : (
            <>
              <ListSearchInput
                value={search}
                onChange={setSearch}
                placeholder="Search commands"
              />
              {filtered!.length === 0 ? (
                <ListSearchEmpty query={search} />
              ) : (
                filtered!.map((command) => (
                  <CommandListEntry command={command!} key={command!.id} />
                ))
              )}
            </>
          )}

          {/* The deploy button is never disabled: deleting a command doesn't
              change any remaining command's updated_at, and deleting the last
              one leaves no command to compare at all. Gating on "has
              undeployed changes" made deleted commands unremovable. */}
          <div className="flex gap-5 justify-between flex-col md:flex-row">
            <div className="flex gap-5 flex-col md:flex-row">
              <CommandCreateDialog>
                <Button>Create command</Button>
              </CommandCreateDialog>
              <FlowImportDialog kind="command">
                <Button variant="outline">Import command</Button>
              </FlowImportDialog>
            </div>

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
