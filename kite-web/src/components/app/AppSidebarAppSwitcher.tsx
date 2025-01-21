import { BotIcon, ChevronsUpDown, CogIcon } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { useApp, useApps } from "@/lib/hooks/api";
import Link from "next/link";
import { useRouter } from "next/router";
import { useCallback } from "react";

export default function AppSidebarAppSwitcher() {
  const { isMobile } = useSidebar();

  const router = useRouter();

  const apps = useApps();
  const app = useApp();

  const setApp = useCallback(
    (appId: string) => {
      router.push({
        pathname: router.pathname,
        query: {
          ...router.query,
          appId,
        },
      });
    },
    [router]
  );

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <BotIcon className="size-5" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">{app?.name}</span>
                <span className="truncate text-xs">Open Beta</span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-xs text-muted-foreground">
              Apps
            </DropdownMenuLabel>
            {apps?.map((app, index) => (
              <DropdownMenuItem
                key={app!.id}
                onClick={() => setApp(app!.id)}
                className="gap-2 p-2"
              >
                <div className="flex size-6 items-center justify-center rounded-sm border">
                  <BotIcon className="size-4 shrink-0" />
                </div>
                {app!.name}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem className="gap-2 p-2 cursor-pointer" asChild>
              <Link href="/apps">
                <div className="flex size-6 items-center justify-center rounded-md border bg-background">
                  <CogIcon className="size-4" />
                </div>
                <div className="font-medium text-muted-foreground">Manage apps</div>
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
