import { useUser } from "@/lib/hooks/api";
import { useEffect } from "react";
import { openPanel, useTrack } from "@/lib/analytics";

// Accounts this new count as signing up on this visit.
const signUpWindowMs = 10 * 60 * 1000;

export default function AnalyticsProvider() {
  const track = useTrack();
  const user = useUser();

  useEffect(() => {
    if (user?.id) {
      openPanel?.then((op) =>
        op.identify({
          profileId: user.id,
          firstName: user.discord_username,
          lastName: user.display_name,
          email: user.email,
        })
      );

      const key = `kite-signed-up-${user.id}`;
      const isNew =
        Date.now() - new Date(user.created_at).getTime() < signUpWindowMs;
      if (isNew && !localStorage.getItem(key)) {
        localStorage.setItem(key, "true");
        track("sign_up");
      }
    }
  }, [track, user]);

  return null;
}
