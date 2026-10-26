import { useApp, useAppPlan, useUser } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useEffect } from "react";
import { openPanel, setAppGroup, updateAppGroup, track } from "@/lib/analytics";

// Accounts this new count as signing up on this visit.
const signUpWindowMs = 10 * 60 * 1000;

export default function AnalyticsProvider() {
  const user = useUser();
  const appId = useAppId();

  useEffect(() => {
    if (user?.id) {
      openPanel?.then((op) =>
        op.identify({
          profileId: user.id,
          firstName: user.discord_username,
          lastName: user.display_name,
          email: user.email ?? undefined,
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
  }, [user]);

  // After identifying, so the user is linked to the app.
  const identified = !!user?.id;
  useEffect(() => setAppGroup(appId, identified), [appId, identified]);

  // Its queries are only needed for analytics.
  return appId && openPanel ? <AppGroup /> : null;
}

function AppGroup() {
  const app = useApp();
  const plan = useAppPlan();

  // The app query refetches while commands deploy, so only the values sent
  // are deps.
  const id = app?.id;
  const name = app?.name;
  const enabled = app?.enabled;
  const createdAt = app?.created_at;
  const guildCount = app?.discord_guild_count ?? 0;
  const userInstallCount = app?.discord_user_install_count ?? 0;
  const planId = plan?.id;

  useEffect(() => {
    if (!id || !name) return;

    updateAppGroup(
      { id, name },
      {
        enabled: enabled!,
        created_at: createdAt!,
        guild_count: guildCount,
        user_install_count: userInstallCount,
        ...(planId && { plan: planId }),
      }
    );
  }, [id, name, enabled, createdAt, guildCount, userInstallCount, planId]);

  return null;
}
