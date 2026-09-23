import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useEventListeners } from "@/lib/hooks/api";
import EventListenerListEntry from "./EventListenerListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import EventListenerCreateDialog from "./EventListenerCreateDialog";
import FlowImportDialog from "./FlowImportDialog";

export default function EventListenerList() {
  const listeners = useEventListeners();

  const listenerActions = (
    <div className="flex gap-5 flex-col md:flex-row">
      <EventListenerCreateDialog>
        <Button>Create event listener</Button>
      </EventListenerCreateDialog>
      <FlowImportDialog kind="event_listener">
        <Button variant="outline">Import event listener</Button>
      </FlowImportDialog>
    </div>
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
          action={listenerActions}
        />
      ) : (
        <>
          {listeners.map((listener, i) => (
            <EventListenerListEntry listener={listener!} key={i} />
          ))}
          {listenerActions}
        </>
      )}
    </AutoAnimate>
  );
}
