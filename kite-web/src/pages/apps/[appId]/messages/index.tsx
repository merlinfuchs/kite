import AppLayout from "@/components/app/AppLayout";
import MessageList from "@/components/app/MessageList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Message Templates",
  },
];

export default function AppMessagesPage() {
  return (
    <AppLayout title="Message Templates" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Message Templates
        </h1>
        <p className="text-muted-foreground text-sm">
          Create message templates that can be used as responses to commands and
          events in your app.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/message`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <MessageList />
    </AppLayout>
  );
}
