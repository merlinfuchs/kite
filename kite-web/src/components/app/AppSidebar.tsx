import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar";
import AppSidebarAppSwitcher from "./AppSidebarAppSwitcher";
import AppSidebarExternalNav from "./AppSidebarExternalNav";
import AppSidebarMainNav from "./AppSidebarMainNav";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";
import { useCallback, useEffect, useLayoutEffect, useRef } from "react";

const useIsomorphicLayoutEffect =
  typeof window !== "undefined" ? useLayoutEffect : useEffect;

let savedSidebarScrollTop = 0;

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const contentRef = useRef<HTMLDivElement>(null);

  useIsomorphicLayoutEffect(() => {
    if (!contentRef.current) return;

    if (savedSidebarScrollTop > 0) {
      contentRef.current.scrollTop = savedSidebarScrollTop;
    } else {
      const activeItem = contentRef.current.querySelector(
        '[data-active="true"]'
      );
      if (activeItem) {
        activeItem.scrollIntoView({ block: "nearest" });
      }
    }
  }, []);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    savedSidebarScrollTop = e.currentTarget.scrollTop;
  }, []);

  return (
    <Sidebar collapsible="icon" variant="floating" {...props}>
      <SidebarHeader>
        <AppSidebarAppSwitcher />
      </SidebarHeader>
      <SidebarContent
        ref={contentRef}
        onScroll={handleScroll}
        className="no-scrollbar"
      >
        <AppSidebarMainNav />
        <AppSidebarStudioNav />
        <AppSidebarExternalNav />
      </SidebarContent>
      <SidebarFooter>
        <AppSidebarUserNav />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
