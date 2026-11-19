import { Button } from "@/components/ui/button";
import { useApp, useAppPlan, useAppStateGuilds } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { cn } from "@/lib/utils";
import { ChevronDownIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { createContext, ReactNode } from "react";
import logo from "@/assets/logo/orange@1024.png";
import ThemeSwitch from "../common/ThemeSwitch";
import UserMenu, { UserAvatar } from "./UserMenu";

// InKiteBotShell tells AppLayout to leave out its header with the sidebar
// toggle and breadcrumbs, the shell has its own.
export const InKiteBotShell = createContext(false);

// KiteBotShell is the shell of apps on the Kite bot, which only add templates.
// It has a few tabs instead of the sidebar, and none while setting up, so
// nothing distracts from picking templates.
export default function KiteBotShell({
  children,
  tabs = true,
}: {
  children: ReactNode;
  tabs?: boolean;
}) {
  const appId = useAppId();
  const app = useApp();
  const plan = useAppPlan();
  const guild = useAppStateGuilds()?.[0];
  const router = useRouter();

  const items = [
    { title: "Templates", url: "/apps/[appId]", exact: true },
    { title: "Logs", url: "/apps/[appId]/logs" },
    {
      // Paid apps manage their plan here, the rest are asked to upgrade.
      title: plan && !plan.default ? "Plan" : "Upgrade",
      url: "/apps/[appId]/premium",
    },
    { title: "Settings", url: "/apps/[appId]/settings" },
  ];

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

      {tabs && (
        <nav className="flex-none overflow-x-auto border-b bg-background px-3">
          <div className="mx-auto flex max-w-3xl gap-1">
            {items.map((item) => {
              const active = item.exact
                ? router.pathname === item.url
                : router.pathname.startsWith(item.url);
              return (
                <Link
                  key={item.url}
                  href={{ pathname: item.url, query: { appId } }}
                  aria-current={active ? "page" : undefined}
                  className={cn(
                    "-mb-px whitespace-nowrap border-b-2 px-3 py-3 text-sm font-medium transition-colors",
                    active
                      ? "border-primary text-foreground"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  )}
                >
                  {item.title}
                </Link>
              );
            })}
          </div>
        </nav>
      )}

      <div className="mx-auto w-full max-w-3xl flex-1 px-5 pb-20 pt-8">
        <InKiteBotShell.Provider value={true}>
          {children}
        </InKiteBotShell.Provider>
      </div>
    </div>
  );
}
