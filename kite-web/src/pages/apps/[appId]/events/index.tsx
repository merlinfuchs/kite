import { getAppShellLayout } from "@/components/app/AppShell";
import EventListenerPage from "@/components/app/EventListenerPage";

export default function AppEventsPage() {
  return <EventListenerPage source="discord" />;
}

AppEventsPage.getLayout = getAppShellLayout;
