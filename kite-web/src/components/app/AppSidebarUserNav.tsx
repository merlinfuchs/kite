import { ChevronsUpDown } from "lucide-react";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { useUser } from "@/lib/hooks/api";
import UserMenu, { UserAvatar } from "./UserMenu";

export default function AppSidebarUserNav() {
  const { isMobile } = useSidebar();
  const user = useUser();

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <UserMenu
          className="w-[--radix-dropdown-menu-trigger-width]"
          side={isMobile ? "bottom" : "right"}
        >
          <SidebarMenuButton
            size="lg"
            className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
          >
            <UserAvatar className="h-8 w-8" />
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-semibold">
                {user?.display_name}
              </span>
              <span className="truncate text-xs">
                {user && `@${user.discord_username}`}
              </span>
            </div>
            <ChevronsUpDown className="ml-auto size-4" />
          </SidebarMenuButton>
        </UserMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
