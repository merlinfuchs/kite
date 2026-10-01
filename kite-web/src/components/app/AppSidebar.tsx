import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import { useRouter } from "next/router";
import { useEffect } from "react";
import AppSidebarAppSwitcher from "./AppSidebarAppSwitcher";
import AppSidebarExternalNav from "./AppSidebarExternalNav";
import AppSidebarMainNav from "./AppSidebarMainNav";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const router = useRouter();
  const { setOpenMobile } = useSidebar();

  // The sidebar stays mounted across pages, so close the mobile drawer after
  // navigating.
  useEffect(() => {
    setOpenMobile(false);
  }, [router.asPath, setOpenMobile]);

  return (
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
  );
}
