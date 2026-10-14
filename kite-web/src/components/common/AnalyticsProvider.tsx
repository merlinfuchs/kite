import { useUser } from "@/lib/hooks/api";
import { OpenPanelComponent, useOpenPanel } from "@openpanel/nextjs";
import { useEffect } from "react";
import { useTrack } from "@/lib/analytics";

// Accounts this new count as signing up on this visit.
const signUpWindowMs = 10 * 60 * 1000;

export default function AnalyticsProvider() {
  const { identify } = useOpenPanel();
  const track = useTrack();
  const user = useUser();

  useEffect(() => {
    if (user?.id) {
      identify({
        profileId: user.id,
        firstName: user.discord_username,
        lastName: user.display_name,
        email: user.email,
      });

      const key = `kite-signed-up-${user.id}`;
      const isNew =
        Date.now() - new Date(user.created_at).getTime() < signUpWindowMs;
      if (isNew && !localStorage.getItem(key)) {
        localStorage.setItem(key, "true");
        track("sign_up");
      }
    }
  }, [identify, track, user]);

  if (process.env.NODE_ENV !== "production" || typeof window === "undefined") {
    return null;
  }

  return (
    <OpenPanelComponent
      clientId="b14a6614-59bb-481e-8de0-ae685dd67de1"
      apiUrl="https://analytics.xenon.bot/api"
      scriptUrl="https://analytics.xenon.bot/op1.js"
      trackScreenViews={true}
      trackOutgoingLinks={true}
    />
  );
}
