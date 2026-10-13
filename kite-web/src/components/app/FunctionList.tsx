import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import {
  useCommands,
  useEventListeners,
  useFunctions,
  useMessages,
} from "@/lib/hooks/api";
import { getFunctionUsages } from "@/lib/flow/functions";
import FunctionListEntry from "./FunctionListEntry";
import FunctionCreateDialog from "./FunctionCreateDialog";
import FlowImportDialog from "./FlowImportDialog";

export default function FunctionList() {
  const functions = useFunctions();
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();

  // Unknown until everything that can call a function is loaded.
  const usagesLoaded = !!(functions && commands && eventListeners && messages);

  const createButton = (
    <div className="flex gap-5 flex-col md:flex-row">
      <FunctionCreateDialog>
        <Button>Create function</Button>
      </FunctionCreateDialog>
      <FlowImportDialog kind="function">
        <Button variant="outline">Import function</Button>
      </FlowImportDialog>
    </div>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!functions ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : functions.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no functions"
          description="Create a function for blocks you use in several flows, then call it from any of them!"
          action={createButton}
        />
      ) : (
        <>
          {functions.map((fn) => (
            <FunctionListEntry
              function={fn!}
              key={fn!.id}
              usages={
                usagesLoaded
                  ? getFunctionUsages(fn!.id, {
                      commands: commands!,
                      eventListeners: eventListeners!,
                      messages: messages!,
                      functions: functions!,
                    })
                  : undefined
              }
            />
          ))}
          <div className="flex">{createButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
