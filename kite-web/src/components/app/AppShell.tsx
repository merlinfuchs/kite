import { AppSidebar } from "@/components/app/AppSidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useApp } from "@/lib/hooks/api";
import { useRouter } from "next/router";
import { ReactElement, ReactNode } from "react";
import { toast } from "sonner";
import AppDisabledPopup from "./AppDisabledPopup";
import AppIntentsPopup from "./AppIntentsPopup";
import UpsellPopup from "./UpsellPopup";

// Rendered from _app via getLayout, so the sidebar stays mounted and keeps its
// open state and scroll position when navigating between app pages.
export default function AppShell({ children }: { children: ReactNode }) {
  const router = useRouter();

  useApp((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load app: ${res?.error.message} (${res?.error.code})`
      );
      if (
        res.error.code === "unknown_app" ||
        res.error.code === "missing_access"
      ) {
        router.push({
          pathname: "/apps",
        });
      }
    }
  });

  // TODO: remember open state of sidebar on desktop
  return (
    <SidebarProvider className="bg-muted/30">
      <AppSidebar />
      <SidebarInset className="bg-transparent min-h-[100dvh] min-w-0 max-w-[1500px] mx-auto">
        {children}

        <AppDisabledPopup />
        <AppIntentsPopup />
        <UpsellPopup />
      </SidebarInset>
    </SidebarProvider>
  );
}

export function getAppShellLayout(page: ReactElement) {
  return <AppShell>{page}</AppShell>;
}
