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
import { useAppId } from "@/lib/hooks/params";
import { cn } from "@/lib/utils";
import {
  LayoutDashboardIcon,
  MenuIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect } from "react";
import AppSidebarAppSwitcher from "./AppSidebarAppSwitcher";
import AppSidebarExternalNav from "./AppSidebarExternalNav";
import AppSidebarMainNav from "./AppSidebarMainNav";
import AppSidebarSupportCard from "./AppSidebarSupportCard";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const router = useRouter();
  const { isMobile, openMobile, setOpenMobile } = useSidebar();
  const appId = useAppId();

  // Close the mobile drawer after navigating or when resizing to desktop.
  useEffect(() => {
    setOpenMobile(false);
  }, [router.asPath, isMobile, setOpenMobile]);

  const isDashboardActive = router.pathname === "/apps/[appId]";
  const isCommandsActive = router.pathname.startsWith("/apps/[appId]/commands");
  const isEventsActive = router.pathname.startsWith("/apps/[appId]/events");

  const isOtherActive =
    !isDashboardActive && !isCommandsActive && !isEventsActive;

  return (
    <>
      {!isMobile && (
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
            <AppSidebarSupportCard />
            <AppSidebarUserNav />
          </SidebarFooter>
          <SidebarRail />
        </Sidebar>
      )}

      {appId && (
        <nav
          aria-label="Mobile Navigation"
          className="z-40 block md:hidden select-none fixed bottom-0 inset-x-0 bg-background/95 backdrop-blur-md border-t border-border shadow-lg pb-[env(safe-area-inset-bottom)]"
        >
          <div className="grid grid-cols-4 items-center mx-auto h-16 max-w-md px-3">
            <Link
              href={{ pathname: "/apps/[appId]", query: { appId } }}
              aria-current={isDashboardActive ? "page" : undefined}
              className={cn(
                "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                isDashboardActive
                  ? "text-primary font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <LayoutDashboardIcon className="size-5" />
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Dashboard
              </span>
            </Link>

            <Link
              href={{ pathname: "/apps/[appId]/commands", query: { appId } }}
              aria-current={isCommandsActive ? "page" : undefined}
              className={cn(
                "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                isCommandsActive
                  ? "text-primary font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <SlashSquareIcon className="size-5" />
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Commands
              </span>
            </Link>

            <Link
              href={{ pathname: "/apps/[appId]/events", query: { appId } }}
              aria-current={isEventsActive ? "page" : undefined}
              className={cn(
                "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                isEventsActive
                  ? "text-primary font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <SatelliteDishIcon className="size-5" />
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Listeners
              </span>
            </Link>

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
              aria-expanded={openMobile}
            >
              <MenuIcon className="size-5" />
              <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                Menu
              </span>
            </button>
          </div>
        </nav>
      )}

      <Drawer
        open={openMobile && isMobile}
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
