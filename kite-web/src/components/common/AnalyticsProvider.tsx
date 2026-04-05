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

  if (process.env.NODE_ENV !== "production" && typeof window === "undefined") {
    return null;
  }

  return (
    <OpenPanelComponent
      clientId="3d379370-3ce9-4a92-b7ea-2c663b7fa7dd"
      apiUrl="https://analytics.vaven.io/api"
      trackScreenViews={true}
      trackOutgoingLinks={true}
    />
  );
}
