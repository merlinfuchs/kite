import {
  VariableIcon,
  LibraryBigIcon,
  SlashSquareIcon,
  type LucideIcon,
  MailPlusIcon,
  SatelliteDishIcon,
} from "lucide-react";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useCallback, useMemo } from "react";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import Link from "next/link";

export default function AppSidebarStudioNav() {
  const appId = useAppId();
  const router = useRouter();

  const isActive = useCallback(
    (path: string, exact = false) => {
      if (exact) {
        return router.pathname === path;
      }

      return router.pathname.startsWith(path);
    },
    [router.pathname]
  );

  const items = useMemo(() => {
    return [
      {
        name: "Commands",
        url: "/apps/[appId]/commands",
        icon: SlashSquareIcon,
        active: isActive("/apps/[appId]/commands"),
      },
      {
        name: "Event Listeners",
        url: "/apps/[appId]/events",
        icon: SatelliteDishIcon,
        active: isActive("/apps/[appId]/events"),
      },
      {
        name: "Message Templates",
        url: "/apps/[appId]/messages",
        icon: MailPlusIcon,
        active: isActive("/apps/[appId]/messages"),
      },
      {
        name: "Stores Variables",
        url: "/apps/[appId]/variables",
        icon: VariableIcon,
        active: isActive("/apps/[appId]/variables"),
      },
      /* {
        name: "Templates",
        url: "/apps/[appId]/templates",
        icon: LibraryBigIcon,
        active: isActive("/apps/[appId]/templates"),
      }, */
    ];
  }, [isActive]);

  return (
    <SidebarGroup className="">
      <SidebarGroupLabel>Studio</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton asChild isActive={item.active}>
              <Link
                href={{
                  pathname: item.url,
                  query: {
                    appId,
                  },
                }}
              >
                <item.icon />
                <span>{item.name}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
