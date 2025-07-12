import {
  ChevronRightIcon,
  CrownIcon,
  LayoutDashboardIcon,
  MessageSquareWarningIcon,
  Settings2Icon,
  TelescopeIcon,
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
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { useRouter } from "next/router";
import { useCallback, useMemo } from "react";
import { usePlugins } from "@/lib/hooks/api";
import DynamicIcon from "../icons/DynamicIcon";

export default function AppSidebarPluginNav() {
  const appId = useAppId();
  const router = useRouter();

  const plugins = usePlugins();

  const isActive = useCallback(
    (path: string, exact = false) => {
      if (exact) {
        return router.asPath === path;
      }

      return router.asPath.startsWith(path);
    },
    [router.asPath]
  );

  const items = useMemo(() => {
    return (
      plugins?.map((plugin) => ({
        title: plugin!.metadata.name,
        url: `/apps/${appId}/plugins/${plugin!.id}`,
        active: isActive(`/apps/${appId}/plugins/${plugin!.id}`),
        icon: plugin!.metadata.icon,
      })) ?? []
    );
  }, [isActive, plugins, appId]);

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Plugins</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton
              tooltip={item.title}
              asChild
              isActive={item.active}
            >
              <Link
                href={{
                  pathname: item.url,
                  query: {
                    appId,
                  },
                }}
              >
                {/* TODO: Check how much DynamicIcon increases bundle size and build time */}
                {item.icon && <DynamicIcon name={item.icon as any} />}
                <span>{item.title}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
