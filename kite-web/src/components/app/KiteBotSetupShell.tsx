import { Button } from "@/components/ui/button";
import { useApp, useAppStateGuilds } from "@/lib/hooks/api";
import Link from "next/link";
import { createContext, ReactNode } from "react";
import logo from "@/assets/logo/orange@1024.png";
import ThemeSwitch from "../common/ThemeSwitch";
import UserMenu, { UserAvatar } from "./UserMenu";

// InKiteBotSetupShell tells AppLayout to leave out its header with the
// sidebar toggle and breadcrumbs, the shell has its own.
export const InKiteBotSetupShell = createContext(false);

// KiteBotSetupShell is the shell of the setup page of apps on the Kite bot.
// It has no navigation, so nothing distracts from picking templates.
export default function KiteBotSetupShell({
  children,
}: {
  children: ReactNode;
}) {
  const app = useApp();
  const guild = useAppStateGuilds()?.[0];

  return (
    <div className="flex min-h-[100dvh] flex-col bg-muted/30">
      <header className="flex h-14 flex-none items-center justify-between gap-3 border-b bg-background px-5">
        <div className="flex min-w-0 items-center gap-3">
          <Link
            href="/apps"
            className="flex flex-none items-center gap-2 rounded-md font-semibold outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <img src={logo.src} className="h-6 w-6" alt="" />
            Kite
          </Link>
          <span className="text-muted-foreground/50">/</span>
          <div className="flex min-w-0 items-center gap-2 text-sm font-medium">
            {guild?.icon_url ? (
              <img
                src={guild.icon_url}
                className="h-6 w-6 flex-none rounded-full"
                alt=""
              />
            ) : (
              <div className="flex h-6 w-6 flex-none items-center justify-center rounded-full bg-muted text-xs">
                {(guild?.name ?? app?.name ?? "").slice(0, 1)}
              </div>
            )}
            <span className="truncate">{guild?.name ?? app?.name}</span>
          </div>
        </div>
        <div className="flex flex-none items-center gap-3">
          <ThemeSwitch className="text-muted-foreground hover:text-foreground" />
          <UserMenu>
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full"
              aria-label="Account menu"
            >
              <UserAvatar className="size-7 text-xs" />
            </Button>
          </UserMenu>
        </div>
      </header>

      <div className="mx-auto w-full max-w-3xl flex-1 px-5 pb-20 pt-8">
        <InKiteBotSetupShell.Provider value={true}>
          {children}
        </InKiteBotSetupShell.Provider>
      </div>
    </div>
  );
}
