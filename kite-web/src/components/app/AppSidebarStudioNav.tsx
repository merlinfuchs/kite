import {
  VariableIcon,
  StoreIcon,
  KeyRoundIcon,
  PlugIcon,
  SlashSquareIcon,
  MailPlusIcon,
  SatelliteDishIcon,
  BlocksIcon,
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

  // What the bot does, and what its flows use.
  const sections = useMemo(() => {
    return [
      {
        label: "Studio",
        items: [
          {
            name: "Commands",
            url: "/apps/[appId]/commands",
            icon: SlashSquareIcon,
          },
          {
            name: "Event Listeners",
            url: "/apps/[appId]/events",
            icon: SatelliteDishIcon,
          },
          {
            name: "Message Templates",
            url: "/apps/[appId]/messages",
            icon: MailPlusIcon,
          },
          {
            name: "Plugins",
            url: "/apps/[appId]/plugins",
            icon: BlocksIcon,
          },
          {
            name: "Marketplace",
            url: "/apps/[appId]/marketplace",
            icon: StoreIcon,
          },
        ],
      },
      {
        label: "Resources",
        items: [
          {
            name: "Stored Variables",
            url: "/apps/[appId]/variables",
            icon: VariableIcon,
          },
          {
            name: "Secrets",
            url: "/apps/[appId]/secrets",
            icon: KeyRoundIcon,
          },
          {
            name: "Integrations",
            url: "/apps/[appId]/integrations",
            icon: PlugIcon,
          },
        ],
      },
    ];
  }, []);

  return (
    <>
      {sections.map((section) => (
        <SidebarGroup key={section.label}>
          <SidebarGroupLabel>{section.label}</SidebarGroupLabel>
          <SidebarMenu>
            {section.items.map((item) => (
              <SidebarMenuItem key={item.name}>
                <SidebarMenuButton asChild isActive={isActive(item.url)}>
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
      ))}
    </>
  );
}
