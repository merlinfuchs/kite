import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Separator } from "@/components/ui/separator";
import { useApps } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useKiteSettings } from "@/lib/hooks/useKiteSettings";
import { cn } from "@/lib/utils";
import {
  LayoutDashboardIcon,
  MenuIcon,
  MessageSquareWarningIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useMemo } from "react";
import AppSidebarAppSwitcher from "./AppSidebarAppSwitcher";
import AppSidebarExternalNav from "./AppSidebarExternalNav";
import AppSidebarMainNav from "./AppSidebarMainNav";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const router = useRouter();
  const { openMobile, setOpenMobile } = useSidebar();
  const appId = useAppId();
  const apps = useApps();
  const effectiveAppId =
    appId || (router.query.appId as string) || apps?.[0]?.id;
  const { settings } = useKiteSettings();

  // The sidebar stays mounted across pages, so close the mobile drawer after navigating.
  useEffect(() => {
    setOpenMobile(false);
  }, [router.asPath, setOpenMobile]);

  const isDashboardActive = router.pathname === "/apps/[appId]";
  const isCommandsActive = router.pathname.startsWith("/apps/[appId]/commands");
  const isEventsActive = router.pathname.startsWith("/apps/[appId]/events");
  const isLogsActive = router.pathname.startsWith("/apps/[appId]/logs");

  const isOtherActive =
    !isDashboardActive &&
    !isCommandsActive &&
    !(settings.mobileNavTabs !== "3-tabs" && isEventsActive) &&
    !(settings.mobileNavTabs === "5-tabs" && isLogsActive);

  const dashboardHref = useMemo(
    () =>
      effectiveAppId
        ? {
            pathname: "/apps/[appId]",
            query: { appId: effectiveAppId },
          }
        : "/apps",
    [effectiveAppId]
  );

  const commandsHref = useMemo(
    () =>
      effectiveAppId
        ? {
            pathname: "/apps/[appId]/commands",
            query: { appId: effectiveAppId },
          }
        : "/apps",
    [effectiveAppId]
  );

  const eventsHref = useMemo(
    () =>
      effectiveAppId
        ? {
            pathname: "/apps/[appId]/events",
            query: { appId: effectiveAppId },
          }
        : "/apps",
    [effectiveAppId]
  );

  const logsHref = useMemo(
    () =>
      effectiveAppId
        ? {
            pathname: "/apps/[appId]/logs",
            query: { appId: effectiveAppId },
          }
        : "/apps",
    [effectiveAppId]
  );

  const isFloating = settings.mobileNavStyle === "floating";
  const showLabels = settings.showMobileLabels;

  const gridColsClass =
    settings.mobileNavTabs === "5-tabs"
      ? "grid-cols-5"
      : settings.mobileNavTabs === "4-tabs"
      ? "grid-cols-4"
      : "grid-cols-3";

  return (
    <>
      <Sidebar collapsible="icon" variant="floating" {...props}>
        <SidebarHeader>
          <AppSidebarAppSwitcher />
        </SidebarHeader>
        <SidebarContent>
          <AppSidebarMainNav />
          <AppSidebarStudioNav />
          <AppSidebarExternalNav />
        </SidebarContent>
        <SidebarFooter>
          <AppSidebarUserNav />
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>

      <nav
        aria-label="Mobile Navigation"
        className={cn(
          "z-40 block md:hidden select-none transition-all duration-200",
          isFloating
            ? "fixed bottom-3 inset-x-3 max-w-sm mx-auto bg-background/95 backdrop-blur-md border border-border/80 shadow-xl rounded-full px-2 py-1.5"
            : "fixed bottom-0 inset-x-0 bg-background/95 backdrop-blur-md border-t border-border shadow-lg pb-[env(safe-area-inset-bottom)]"
        )}
      >
        <div
          className={cn(
            "grid items-center mx-auto",
            isFloating ? "h-14 max-w-sm px-1" : "h-16 max-w-md px-3",
            gridColsClass
          )}
        >
          <Link
            href={dashboardHref}
            className={cn(
              "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
              isDashboardActive
                ? "text-primary font-semibold"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <LayoutDashboardIcon className="size-5" />
            {showLabels && (
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Dashboard
              </span>
            )}
          </Link>

          <Link
            href={commandsHref}
            className={cn(
              "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
              isCommandsActive
                ? "text-primary font-semibold"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <SlashSquareIcon className="size-5" />
            {showLabels && (
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Commands
              </span>
            )}
          </Link>

          {settings.mobileNavTabs !== "3-tabs" && (
            <Link
              href={eventsHref}
              className={cn(
                "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                isEventsActive
                  ? "text-primary font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <SatelliteDishIcon className="size-5" />
              {showLabels && (
                <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                  Listeners
                </span>
              )}
            </Link>
          )}

          {settings.mobileNavTabs === "5-tabs" && (
            <Link
              href={logsHref}
              className={cn(
                "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                isLogsActive
                  ? "text-primary font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <MessageSquareWarningIcon className="size-5" />
              {showLabels && (
                <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                  Logs
                </span>
              )}
            </Link>
          )}

          <button
            type="button"
            onClick={() => setOpenMobile(!openMobile)}
            className={cn(
              "flex flex-col items-center justify-center py-1 rounded-lg transition-colors cursor-pointer",
              openMobile || isOtherActive
                ? "text-primary font-semibold"
                : "text-muted-foreground hover:text-foreground"
            )}
            aria-label="Open navigation menu"
          >
            <MenuIcon className="size-5" />
            {showLabels && (
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Menu
              </span>
            )}
          </button>
        </div>
      </nav>

      <Drawer
        open={openMobile}
        onOpenChange={setOpenMobile}
        shouldScaleBackground={false}
      >
        <DrawerContent
          className={cn(
            "fixed inset-x-0 bottom-0 z-50 flex max-h-[85vh] flex-col rounded-t-[16px] border border-b-0 bg-sidebar text-sidebar-foreground shadow-2xl focus:outline-none",
            "after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-sidebar"
          )}
        >
          <DrawerHeader className="sr-only">
            <DrawerTitle>Navigation Menu</DrawerTitle>
            <DrawerDescription>
              Mobile navigation options for Kite
            </DrawerDescription>
          </DrawerHeader>

          <div className="px-3 pt-2 pb-1 shrink-0">
            <AppSidebarAppSwitcher />
          </div>

          <Separator className="my-1 bg-sidebar-border" />

          <div className="flex-1 overflow-y-auto px-2 space-y-2 pb-2 pt-1">
            <AppSidebarMainNav />
            <AppSidebarStudioNav />
            <AppSidebarExternalNav />
          </div>

          <div className="border-t border-sidebar-border bg-sidebar px-3 py-2 shrink-0">
            <AppSidebarUserNav />
          </div>
        </DrawerContent>
      </Drawer>
    </>
  );
}
