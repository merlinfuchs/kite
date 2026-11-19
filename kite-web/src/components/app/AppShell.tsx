import { AppSidebar } from "@/components/app/AppSidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useApp } from "@/lib/hooks/api";
import { useRouter } from "next/router";
import { ReactElement, ReactNode } from "react";
import { toast } from "sonner";
import AppCreditsPopup from "./AppCreditsPopup";
import AppDisabledPopup from "./AppDisabledPopup";
import UpsellPopup from "./UpsellPopup";
import LegalUpdateDialog from "../common/LegalUpdateDialog";
import KiteBotSetupShell from "./KiteBotSetupShell";

// Rendered from _app via getLayout, so the sidebar stays mounted and keeps its
// open state and scroll position when navigating between app pages.
export default function AppShell({ children }: { children: ReactNode }) {
  const router = useRouter();

  const app = useApp((res) => {
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

  // The setup page of apps on the Kite bot has no sidebar. People who just
  // added the bot land on it, so it doesn't wait for the app with the sidebar.
  const setup = router.pathname === "/apps/[appId]/setup";
  if (setup && (!app || app.bot_type === "shared")) {
    return (
      <KiteBotSetupShell>
        {children}

        <AppDisabledPopup />
        <AppCreditsPopup />
        <LegalUpdateDialog />
      </KiteBotSetupShell>
    );
  }

  // TODO: remember open state of sidebar on desktop
  return (
    <SidebarProvider className="bg-muted/30">
      <AppSidebar />
      <SidebarInset className="bg-transparent min-h-[100dvh] min-w-0 max-w-[1500px] mx-auto pb-20 md:pb-0">
        {children}

        <AppDisabledPopup />
        <AppCreditsPopup />
        <UpsellPopup />
        <LegalUpdateDialog />
      </SidebarInset>
    </SidebarProvider>
  );
}

export function getAppShellLayout(page: ReactElement) {
  return <AppShell>{page}</AppShell>;
}
