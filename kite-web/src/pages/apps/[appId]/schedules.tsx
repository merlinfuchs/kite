import { getAppShellLayout } from "@/components/app/AppShell";
import EventListenerPage from "@/components/app/EventListenerPage";

export default function AppScheduledTasksPage() {
  return <EventListenerPage source="schedule" />;
}

AppScheduledTasksPage.getLayout = getAppShellLayout;
