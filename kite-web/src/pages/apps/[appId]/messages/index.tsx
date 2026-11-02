import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import MessageList from "@/components/app/MessageList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Message templates",
  },
];

export default function AppMessagesPage() {
  return (
    <AppLayout title="Message templates" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Message templates
        </h1>
        <p className="text-muted-foreground text-sm">
          Create message templates that can be used as responses to commands and
          events in your app.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/message`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn more
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <MessageList />
    </AppLayout>
  );
}

AppMessagesPage.getLayout = getAppShellLayout;
