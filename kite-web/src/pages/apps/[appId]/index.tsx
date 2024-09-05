import AppLayout from "@/components/app/AppLayout";
import LogEntryList from "@/components/app/LogEntryList";
import { useApp, useCommands, useMessages } from "@/lib/hooks/api";
import AppInfoCard from "@/components/app/AppInfoCard";
import AppResourceCard from "@/components/app/AppResourceCard";

export default function AppPage() {
  const app = useApp();

  const commands = useCommands();
  const messages = useMessages();

  return (
    <AppLayout>
      <main className="grid flex-1 items-start gap-4 sm:py-0 md:gap-8 lg:grid-cols-3">
        <div className="grid auto-rows-max items-start gap-4 md:gap-8 lg:col-span-2 order-2 lg:order-1">
          <div className="grid gap-4 grid-cols-1 lg:grid-cols-3">
            <AppResourceCard
              title="Commands"
              count={commands?.length || 0}
              actionTitle="Manage commands"
              actionHref={{
                pathname: "/apps/[appId]/commands",
                query: { appId: app?.id },
              }}
            />
            <AppResourceCard
              title="Events"
              count={0}
              actionTitle="Manage events"
              actionHref={{
                pathname: "/apps/[appId]/events",
                query: { appId: app?.id },
              }}
            />
            <AppResourceCard
              title="Messages"
              count={messages?.length || 0}
              actionTitle="Manage messages"
              actionHref={{
                pathname: "/apps/[appId]/messages",
                query: { appId: app?.id },
              }}
            />
          </div>
          <LogEntryList />
        </div>
        <div className="order-1 lg:order-2">
          <AppInfoCard />
        </div>
      </main>
    </AppLayout>
  );
}
