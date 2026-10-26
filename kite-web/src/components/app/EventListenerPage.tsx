import AppLayout from "@/components/app/AppLayout";
import EventListenerList from "@/components/app/EventListenerList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { EventListenerSource, eventSourceName } from "@/lib/flow/eventTypes";

const pageTexts: Record<
  EventListenerSource,
  { description: string; docsAnchor: string }
> = {
  discord: {
    description:
      "Listen for events in the servers your bot is in and take actions based on them.",
    docsAnchor: "",
  },
  schedule: {
    description: "Run flows on a schedule, like every hour or every day.",
    docsAnchor: "#scheduled-event-listeners",
  },
  webhook: {
    description:
      "Run flows when another service, like GitHub, sends a request to a URL.",
    docsAnchor: "#webhook-event-listeners",
  },
};

// EventListenerPage lists the event listeners of one source. Each source has
// its own page in the sidebar.
export default function EventListenerPage({
  source,
}: {
  source: EventListenerSource;
}) {
  const title = eventSourceName(source).plural;
  const texts = pageTexts[source];

  return (
    <AppLayout title={title} breadcrumbs={[{ label: title }]}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">{title}</h1>
        <p className="text-muted-foreground text-sm">
          {texts.description}{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/event${texts.docsAnchor}`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <EventListenerList source={source} />
    </AppLayout>
  );
}
