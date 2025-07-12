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
import AppSidebarPluginNav from "./AppSidebarPluginNav";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" variant="floating" {...props}>
      <SidebarHeader>
        <AppSidebarAppSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <AppSidebarMainNav />
        <AppSidebarStudioNav />
        <AppSidebarPluginNav />
        <AppSidebarExternalNav />
      </SidebarContent>
      <SidebarFooter>
        <AppSidebarUserNav />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
