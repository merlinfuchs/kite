import { getAppShellLayout } from "@/components/app/AppShell";
import EventListenerPage from "@/components/app/EventListenerPage";

export default function AppWebhooksPage() {
  return <EventListenerPage source="webhook" />;
}

AppWebhooksPage.getLayout = getAppShellLayout;
