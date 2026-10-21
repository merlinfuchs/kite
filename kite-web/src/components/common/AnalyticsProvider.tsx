import { useApp, useAppPlan, useUser } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useEffect } from "react";
import {
  clearAppGroup,
  openPanel,
  setAppGroup,
  useTrack,
} from "@/lib/analytics";

// Accounts this new count as signing up on this visit.
const signUpWindowMs = 10 * 60 * 1000;

export default function AnalyticsProvider() {
  const track = useTrack();
  const user = useUser();
  const appId = useAppId();

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

  return appId ? <AppGroup key={appId} /> : null;
}

function AppGroup() {
  const app = useApp();
  const plan = useAppPlan();

  // The app query refetches while commands deploy, so only values are deps.
  const id = app?.id;
  const name = app?.name;
  const enabled = app?.enabled;
  const createdAt = app?.created_at;
  const guildCount = app?.discord_guild_count;
  const userInstallCount = app?.discord_user_install_count;
  const planId = plan?.id;

  useEffect(() => {
    if (!id || !name) return;

    setAppGroup(
      { id, name },
      {
        enabled: enabled!,
        created_at: createdAt!,
        guild_count: guildCount ?? 0,
        user_install_count: userInstallCount ?? 0,
        ...(planId && { plan: planId }),
      }
    );
  }, [id, name, enabled, createdAt, guildCount, userInstallCount, planId]);

  useEffect(() => clearAppGroup, []);

  return null;
}
