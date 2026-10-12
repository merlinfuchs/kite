import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useEventListeners } from "@/lib/hooks/api";
import EventListenerListEntry from "./EventListenerListEntry";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import EventListenerCreateDialog, {
  EventListenerSource,
} from "./EventListenerCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useMemo } from "react";
import { eventSourceName } from "@/lib/flow/eventTypes";

const emptyDescriptions: Record<EventListenerSource, string> = {
  discord: "Run a flow when something happens in the servers your bot is in.",
  schedule: "Run a flow every few minutes, hours or days.",
  webhook: "Run a flow whenever another service sends a request to its URL.",
};

export default function EventListenerList({
  source,
}: {
  source: EventListenerSource;
}) {
  const allListeners = useEventListeners();
  const listeners = useMemo(
    () => allListeners?.filter((l) => l?.source === source),
    [allListeners, source]
  );

  const name = eventSourceName(source);

  const listenerActions = (
    <div className="flex gap-5 flex-col md:flex-row">
      <EventListenerCreateDialog source={source}>
        <Button>Create {name.singular}</Button>
      </EventListenerCreateDialog>
      <FlowImportDialog kind="event_listener" eventSource={source}>
        <Button variant="outline">Import {name.singular}</Button>
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
          title={`There are no ${name.singular}s`}
          description={emptyDescriptions[source]}
          action={listenerActions}
        />
      ) : (
        <>
          {listeners.map((listener) => (
            <EventListenerListEntry listener={listener!} key={listener!.id} />
          ))}
          {listenerActions}
        </>
      )}
    </AutoAnimate>
  );
}
