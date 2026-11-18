import { BotIcon, ChevronsUpDown, CogIcon } from "lucide-react";

import LoadingAvatar from "@/components/common/LoadingAvatar";
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
import { useAppStateUserQuery } from "@/lib/api/queries";
import { useApp, useAppPlan, useApps, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { useRouter } from "next/router";
import { useCallback } from "react";

export default function AppSidebarAppSwitcher() {
  const { isMobile } = useSidebar();

  const router = useRouter();
  const appId = useAppId();

  const apps = useApps();
  const app = useApp();
  const plan = useAppPlan();

  const botUserQuery = useAppStateUserQuery(
    app?.enabled === false ? "" : app?.id ?? appId ?? ""
  );
  const botUser = useResponseData(botUserQuery);
  const botAvatarLoading =
    app?.enabled !== false && (!appId || botUserQuery.isPending);

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
              <LoadingAvatar
                src={botUser?.avatar_url || undefined}
                loading={botAvatarLoading}
                className="aspect-square size-8"
                fallbackClassName="bg-sidebar-primary text-sidebar-primary-foreground"
              >
                <BotIcon className="size-5" aria-hidden="true" />
              </LoadingAvatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">{app?.name}</span>
                <span className="truncate text-xs">{plan?.title}</span>
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
            {apps?.map((item) =>
              item ? (
                <DropdownMenuItem
                  key={item.id}
                  onClick={() => setApp(item.id)}
                  className="gap-2 p-2 cursor-pointer"
                >
                  <div className="flex size-6 items-center justify-center rounded-sm border">
                    <BotIcon className="size-4 shrink-0" aria-hidden="true" />
                  </div>
                  <span className="truncate">{item.name}</span>
                </DropdownMenuItem>
              ) : null
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem className="gap-2 p-2 cursor-pointer" asChild>
              <Link href="/apps">
                <div className="flex size-6 items-center justify-center rounded-md border bg-background">
                  <CogIcon className="size-4" />
                </div>
                <div className="font-medium text-muted-foreground">
                  Manage apps
                </div>
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
