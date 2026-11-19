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
import AppSidebarMainNav, {
  NavItem,
  useMainNavItems,
} from "./AppSidebarMainNav";
import AppSidebarSupportCard from "./AppSidebarSupportCard";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";
import { eventSourceName, eventSources } from "@/lib/flow/eventTypes";
import { useSharedBot } from "@/lib/hooks/api";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const router = useRouter();
  const { isMobile, openMobile, setOpenMobile } = useSidebar();
  const appId = useAppId();
  // Apps on the Kite bot only add templates, so they only get a few pages.
  const sharedBot = useSharedBot();

  // Close the mobile drawer after navigating or when resizing to desktop.
  useEffect(() => {
    setOpenMobile(false);
  }, [router.asPath, isMobile, setOpenMobile]);

  const mainItems = useMainNavItems();
  // The links of the bar at the bottom on phones, the rest is in the menu.
  const mobileItems: Pick<NavItem, "title" | "url" | "icon" | "active">[] =
    sharedBot
      ? mainItems.filter((item) => item.url !== "/apps/[appId]/premium")
      : [
          {
            title: "Dashboard",
            url: "/apps/[appId]",
            icon: LayoutDashboardIcon,
            active: router.pathname === "/apps/[appId]",
          },
          {
            title: "Commands",
            url: "/apps/[appId]/commands",
            icon: SlashSquareIcon,
            active: router.pathname.startsWith("/apps/[appId]/commands"),
          },
          {
            title: "Listeners",
            url: "/apps/[appId]/events",
            icon: SatelliteDishIcon,
            active: eventSources.some((source) =>
              router.pathname.startsWith(eventSourceName(source).path)
            ),
          },
        ];

  const isOtherActive = !mobileItems.some((item) => item.active);

  return (
    <>
      {!isMobile && (
        <Sidebar collapsible="icon" variant="floating" {...props}>
          <SidebarHeader>
            <AppSidebarAppSwitcher />
          </SidebarHeader>
          <SidebarContent>
            <AppSidebarMainNav />
            {!sharedBot && (
              <>
                <AppSidebarStudioNav />
                <AppSidebarExternalNav />
              </>
            )}
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
          aria-label="Mobile navigation"
          className="z-40 block md:hidden select-none fixed bottom-0 inset-x-0 bg-background/95 backdrop-blur-md border-t border-border shadow-lg pb-[env(safe-area-inset-bottom)]"
        >
          <div className="grid grid-cols-4 items-center mx-auto h-16 max-w-md px-3">
            {mobileItems.map((item) => (
              <Link
                key={item.url}
                href={{ pathname: item.url, query: { appId } }}
                aria-current={item.active ? "page" : undefined}
                className={cn(
                  "flex flex-col items-center justify-center py-1 rounded-lg transition-colors",
                  item.active
                    ? "text-primary font-semibold"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                <item.icon className="size-5" />
                <span className="text-[10px] sm:text-[11px] leading-none mt-1">
                  {item.title}
                </span>
              </Link>
            ))}

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
            <DrawerTitle>Navigation menu</DrawerTitle>
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
            {!sharedBot && (
              <>
                <AppSidebarStudioNav />
                <AppSidebarExternalNav />
              </>
            )}
          </div>

          <div className="border-t border-sidebar-border bg-sidebar px-3 py-2 shrink-0">
            <AppSidebarUserNav />
          </div>
        </DrawerContent>
      </Drawer>
    </>
  );
}
