import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useEventListeners } from "@/lib/hooks/api";
import EventListenerListEntry from "./EventListenerListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import EventListenerCreateDialog from "./EventListenerCreateDialog";
import { EventListenerImportDialog } from "./EventListenerImportDialog";

export default function EventListenerList() {
  const listeners = useEventListeners();

  const listenerCreateButton = (
    <EventListenerCreateDialog>
      <Button>Create event listener</Button>
    </EventListenerCreateDialog>
  );

  const listenerImportButton = (
    <EventListenerImportDialog>
      <Button>Import event listener</Button>
    </EventListenerImportDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!listeners ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : listeners.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no event listeners"
          description="You can start now by creating the first event listener!"
          action={
            <div className="flex gap-5 flex-col md:flex-row">
              {listenerCreateButton}
              {listenerImportButton}
            </div>
          }
        />
      ) : (
        <>
          {listeners.map((listener, i) => (
            <EventListenerListEntry listener={listener!} key={i} />
          ))}
          <div className="flex gap-5 flex-col md:flex-row">
            {listenerCreateButton}
            {listenerImportButton}
          </div>
        </>
      )}
    </AutoAnimate>
  );
}