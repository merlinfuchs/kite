import { useState } from "react";
import ListSearchInput, { ListSearchEmpty } from "./ListSearchInput";
import { matchesSearch } from "@/lib/search";
import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useCommands } from "@/lib/hooks/api";
import CommandDeployStatus from "./CommandDeployStatus";
import { SquareSlashIcon } from "lucide-react";
import Link from "next/link";
import { useAppId } from "@/lib/hooks/params";

export default function CommandList() {
  const commands = useCommands();
  const appId = useAppId();
  const [search, setSearch] = useState("");

  const filtered = commands?.filter((command) =>
    matchesSearch(search, [command!.name, command!.description])
  );

  const commandActions = (
    <div className="flex gap-5 flex-col md:flex-row">
      <CommandCreateDialog>
        <Button>Create command</Button>
      </CommandCreateDialog>
      <FlowImportDialog kind="command">
        <Button variant="outline">Import command</Button>
      </FlowImportDialog>
    </div>
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
              action={commandActions}
              footer={
                <>
                  Or{" "}
                  <Link
                    href={{
                      pathname: "/apps/[appId]/templates",
                      query: { appId },
                    }}
                    className="underline hover:text-foreground"
                  >
                    start from a template
                  </Link>
                  .
                </>
              }
            />
          ) : (
            <>
              {commandActions}

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
        </>
      )}
    </AutoAnimate>
  );
}
