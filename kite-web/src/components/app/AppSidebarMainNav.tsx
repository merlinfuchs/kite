import {
  BlocksIcon,
  ChevronRightIcon,
  CrownIcon,
  LayoutDashboardIcon,
  MessageSquareWarningIcon,
  Settings2Icon,
  TelescopeIcon,
  type LucideIcon,
} from "lucide-react";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";
import { useAppPlan, usePluginInstances, useSharedBot } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { track } from "@/lib/analytics";
import { cn } from "@/lib/utils";
import { useRouter } from "next/router";
import { useCallback, useMemo } from "react";

export interface NavItem {
  title: string;
  url: string;
  icon: LucideIcon;
  active: boolean;
  upsell?: boolean;
  items?: { title: string; url: string; active: boolean }[];
}

// useMainNavItems are the app's pages in the "My app" group. Apps on the Kite
// bot only add templates, so they only get a few pages.
export function useMainNavItems(): NavItem[] {
  const router = useRouter();
  // Plugins are replaced by templates, only apps that use one still see them.
  const hasPlugins = !!usePluginInstances()?.length;
  const plan = useAppPlan();
  const sharedBot = useSharedBot();

  const isActive = useCallback(
    (path: string, exact = false) => {
      if (exact) {
        return router.pathname === path;
      }

      return router.pathname.startsWith(path);
    },
    [router.pathname]
  );

  return useMemo((): NavItem[] => {
    const upgrade = {
      // Paid apps manage their plan here, the rest are asked to upgrade.
      title: plan && !plan.default ? "Plan" : "Upgrade",
      url: "/apps/[appId]/premium",
      upsell: !!plan?.default,
      icon: CrownIcon,
      active: isActive("/apps/[appId]/premium"),
    };

    if (sharedBot) {
      return [
        {
          title: "Dashboard",
          url: "/apps/[appId]",
          icon: LayoutDashboardIcon,
          active: isActive("/apps/[appId]", true),
        },
        {
          title: "Logs",
          url: "/apps/[appId]/logs",
          icon: MessageSquareWarningIcon,
          active: isActive("/apps/[appId]/logs"),
        },
        upgrade,
        {
          title: "Settings",
          url: "/apps/[appId]/settings",
          icon: Settings2Icon,
          active: isActive("/apps/[appId]/settings"),
        },
      ];
    }

    return [
      {
        title: "Dashboard",
        url: "/apps/[appId]",
        icon: LayoutDashboardIcon,
        active: isActive("/apps/[appId]", true),
      },
      upgrade,
      {
        title: "Settings",
        url: "/apps/[appId]/settings",
        icon: Settings2Icon,
        active: isActive("/apps/[appId]/settings"),
      },
      ...(hasPlugins
        ? [
            {
              title: "Plugins",
              url: "/apps/[appId]/plugins",
              icon: BlocksIcon,
              active: isActive("/apps/[appId]/plugins"),
            },
          ]
        : []),
      {
        title: "Logs",
        url: "/apps/[appId]/logs",
        icon: MessageSquareWarningIcon,
        active: isActive("/apps/[appId]/logs"),
      },
      {
        title: "Explore",
        url: "#",
        icon: TelescopeIcon,
        active:
          isActive("/apps/[appId]/guilds") || isActive("/apps/[appId]/emojis"),
        items: [
          {
            title: "Servers",
            url: "/apps/[appId]/guilds",
            active: isActive("/apps/[appId]/guilds"),
          },
          {
            title: "Emojis",
            url: "/apps/[appId]/emojis",
            active: isActive("/apps/[appId]/emojis"),
          },
        ],
      },
    ];
  }, [isActive, hasPlugins, plan, sharedBot]);
}

export default function AppSidebarMainNav() {
  const appId = useAppId();
  const items = useMainNavItems();

  return (
    <SidebarGroup>
      <SidebarGroupLabel>My app</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) =>
          item.items ? (
            <Collapsible
              key={item.title}
              asChild
              defaultOpen={item.active}
              className="group/collapsible"
            >
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton tooltip={item.title}>
                    {item.icon && <item.icon />}
                    <span>{item.title}</span>
                    <ChevronRightIcon className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {item.items?.map((subItem) => (
                      <SidebarMenuSubItem key={subItem.title}>
                        <SidebarMenuSubButton asChild isActive={subItem.active}>
                          <Link
                            href={{
                              pathname: subItem.url,
                              query: {
                                appId,
                              },
                            }}
                          >
                            <span>{subItem.title}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          ) : (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                tooltip={item.title}
                asChild
                isActive={item.active}
                className={cn(
                  item.upsell &&
                    "text-primary hover:text-primary active:text-primary data-[active=true]:text-primary"
                )}
              >
                <Link
                  href={{
                    pathname: item.url,
                    query: {
                      appId,
                    },
                  }}
                  onClick={
                    item.upsell
                      ? () => track("upsell_clicked", { from: "sidebar" })
                      : undefined
                  }
                >
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )
        )}
      </SidebarMenu>
    </SidebarGroup>
  );
}
