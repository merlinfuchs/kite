import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useEventListeners } from "@/lib/hooks/api";
import EventListenerListEntry from "./EventListenerListEntry";
import AppEmptyPlaceholder, {
  AppEmptyTemplatesLink,
} from "./AppEmptyPlaceholder";
import EventListenerCreateDialog from "./EventListenerCreateDialog";
import FlowImportDialog from "./FlowImportDialog";
import { useMemo } from "react";
import { EventListenerSource, eventSourceName } from "@/lib/flow/eventTypes";
import {
  ClockIcon,
  LucideIcon,
  SatelliteDishIcon,
  WebhookIcon,
} from "lucide-react";
import SearchableList, { ListActions } from "./SearchableList";

const emptyStates: Record<
  EventListenerSource,
  { icon: LucideIcon; description: string }
> = {
  discord: {
    icon: SatelliteDishIcon,
    description:
      "Run a flow when something happens in the servers your bot is in.",
  },
  schedule: {
    icon: ClockIcon,
    description: "Run a flow every few minutes, hours or days.",
  },
  webhook: {
    icon: WebhookIcon,
    description:
      "Run a flow whenever another service sends a request to its URL.",
  },
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
    <>
      <EventListenerCreateDialog source={source}>
        <Button>Create {name.singular}</Button>
      </EventListenerCreateDialog>
      <FlowImportDialog kind="event_listener" eventSource={source}>
        <Button variant="outline">Import {name.singular}</Button>
      </FlowImportDialog>
    </>
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
          icon={emptyStates[source].icon}
          title={`No ${name.singular}s yet`}
          description={emptyStates[source].description}
          action={<ListActions>{listenerActions}</ListActions>}
          footer={<AppEmptyTemplatesLink />}
        />
      ) : (
        <SearchableList
          items={listeners}
          fields={(listener) => [
            listener.type,
            listener.description,
            listener.source,
          ]}
          renderItem={(listener) => (
            <EventListenerListEntry listener={listener} key={listener.id} />
          )}
          placeholder={`Search ${name.singular}s`}
          actions={listenerActions}
        />
      )}
    </AutoAnimate>
  );
}
