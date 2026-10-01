import { AppSidebar } from "@/components/app/AppSidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { ReactElement, ReactNode } from "react";
import AppDisabledPopup from "./AppDisabledPopup";
import UpsellPopup from "./UpsellPopup";

// Rendered from _app via getLayout, so the sidebar stays mounted and keeps its
// open state and scroll position when navigating between app pages.
export default function AppShell({ children }: { children: ReactNode }) {
  // TODO: remember open state of sidebar on desktop
  return (
    <SidebarProvider className="bg-muted/30">
      <AppSidebar />
      <SidebarInset className="bg-transparent min-h-[100dvh] min-w-0 max-w-[1500px] mx-auto">
        {children}

        <AppDisabledPopup />
        <UpsellPopup />
      </SidebarInset>
    </SidebarProvider>
  );
}

export function getAppShellLayout(page: ReactElement) {
  return <AppShell>{page}</AppShell>;
}
