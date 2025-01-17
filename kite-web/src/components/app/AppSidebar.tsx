import {
  AudioWaveform,
  Bot,
  Command,
  GalleryVerticalEnd,
  LibraryBigIcon,
  MailPlusIcon,
  MessageSquareWarningIcon,
  SatelliteDishIcon,
  Settings2Icon,
  SlashSquareIcon,
  TelescopeIcon,
  VariableIcon,
} from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar";
import AppSidebarAppSwitcher from "./AppSidebarAppSwitcher";
import AppSidebarMainNav from "./AppSidebarMainNav";
import AppSidebarStudioNav from "./AppSidebarStudioNav";
import AppSidebarUserNav from "./AppSidebarUserNav";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" variant="floating" {...props}>
      <SidebarHeader>
        <AppSidebarAppSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <AppSidebarMainNav />
        <AppSidebarStudioNav />
      </SidebarContent>
      <SidebarFooter>
        <AppSidebarUserNav />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
