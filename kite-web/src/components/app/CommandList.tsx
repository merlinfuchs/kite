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
import { SquareSlashIcon } from "lucide-react";
import Link from "next/link";
import { useAppId } from "@/lib/hooks/params";

export default function CommandList() {
  const commands = useCommands();
  const appId = useAppId();

  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

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
            commands.map((command, i) => (
              <CommandListEntry command={command!} key={i} />
            ))
          )}

          {/* The deploy button is never disabled: deleting a command doesn't
              change any remaining command's updated_at, and deleting the last
              one leaves no command to compare at all. Gating on "has
              undeployed changes" made deleted commands unremovable. */}
          <div className="flex gap-5 justify-between items-center flex-col md:flex-row">
            {commands.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                Deleted commands still showing up in Discord? Deploy to remove
                them.
              </p>
            ) : (
              commandActions
            )}

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
