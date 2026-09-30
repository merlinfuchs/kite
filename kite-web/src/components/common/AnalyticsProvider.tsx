import { useUser } from "@/lib/hooks/api";
import { OpenPanelComponent, useOpenPanel } from "@openpanel/nextjs";
import { useEffect } from "react";

export default function AnalyticsProvider() {
  const op = useOpenPanel();
  const user = useUser();

  useEffect(() => {
    if (user?.id) {
      op.identify({
        profileId: user.id,
        firstName: user.discord_username,
        lastName: user.display_name,
        email: user.email,
      });
    }
  }, [user?.id, op.identify]);

  if (process.env.NODE_ENV !== "production" || typeof window === "undefined") {
    return null;
  }

  return (
    <OpenPanelComponent
      clientId="b14a6614-59bb-481e-8de0-ae685dd67de1"
      apiUrl="https://analytics.xenon.bot/api"
      trackScreenViews={true}
      trackOutgoingLinks={true}
    />
  );
}
