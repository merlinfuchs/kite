import { Button } from "../ui/button";
import CommandListEntry from "./CommandListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import CommandCreateDialog from "./CommandCreateDialog";
import { useCommands } from "@/lib/hooks/api";

export default function CommandList() {
  const commands = useCommands();

  const cmdCreateButton = (
    <CommandCreateDialog>
      <Button>Create command</Button>
    </CommandCreateDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!commands ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : commands.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no commands"
          description="You can start now by creating the first command!"
          action={cmdCreateButton}
        />
      ) : (
        <>
          {commands.map((command, i) => (
            <CommandListEntry command={command!} key={i} />
          ))}
          <div className="flex">{cmdCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
