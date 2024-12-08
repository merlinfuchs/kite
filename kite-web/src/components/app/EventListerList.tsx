import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useEventListeners } from "@/lib/hooks/api";
import EventListenerListEntry from "./EventListenerListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import EventListenerCreateDialog from "./EventListenerCreateDialog";

export default function EventListenerList() {
  const listeners = useEventListeners();

  const listenerCreateButton = (
    <EventListenerCreateDialog>
      <Button>Create event listener</Button>
    </EventListenerCreateDialog>
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
          action={listenerCreateButton}
        />
      ) : (
        <>
          {listeners.map((listener, i) => (
            <EventListenerListEntry listener={listener!} key={i} />
          ))}
          <div className="flex">{listenerCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
