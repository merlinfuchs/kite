import {
  VariableIcon,
  KeyRoundIcon,
  PlugIcon,
  LibraryBigIcon,
  SlashSquareIcon,
  MailPlusIcon,
  SatelliteDishIcon,
  BlocksIcon,
  ChevronRightIcon,
  LucideIcon,
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
import { useCallback, useMemo } from "react";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import Link from "next/link";
import { useEventListener } from "@/lib/hooks/api";
import { eventListenerListPath } from "@/lib/flow/eventTypes";

interface NavItem {
  name: string;
  url: string;
  icon: LucideIcon;
  items?: { name: string; url: string; active: boolean }[];
}

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

  // Only loaded on the page of an event listener, to highlight the sub page
  // it belongs to.
  const listener = useEventListener();
  const eventsPath = useMemo(() => {
    if (router.pathname === "/apps/[appId]/events/[eventId]") {
      return listener ? eventListenerListPath(listener.source) : undefined;
    }
    return router.pathname;
  }, [router.pathname, listener]);

  // What the bot does, and what its flows use.
  const sections = useMemo((): { label: string; items: NavItem[] }[] => {
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
            items: [
              {
                name: "Discord Events",
                url: "/apps/[appId]/events",
                active: eventsPath === "/apps/[appId]/events",
              },
              {
                name: "Scheduled Tasks",
                url: "/apps/[appId]/events/schedules",
                active: eventsPath === "/apps/[appId]/events/schedules",
              },
              {
                name: "Webhooks",
                url: "/apps/[appId]/events/webhooks",
                active: eventsPath === "/apps/[appId]/events/webhooks",
              },
            ],
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
            name: "Templates",
            url: "/apps/[appId]/templates",
            icon: LibraryBigIcon,
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
  }, [eventsPath]);

  return (
    <>
      {sections.map((section) => (
        <SidebarGroup key={section.label}>
          <SidebarGroupLabel>{section.label}</SidebarGroupLabel>
          <SidebarMenu>
            {section.items.map((item) =>
              item.items ? (
                <Collapsible
                  key={item.name}
                  asChild
                  defaultOpen={isActive(item.url)}
                  className="group/collapsible"
                >
                  <SidebarMenuItem>
                    <CollapsibleTrigger asChild>
                      <SidebarMenuButton tooltip={item.name}>
                        <item.icon />
                        <span>{item.name}</span>
                        <ChevronRightIcon className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                      </SidebarMenuButton>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {item.items.map((subItem) => (
                          <SidebarMenuSubItem key={subItem.name}>
                            <SidebarMenuSubButton
                              asChild
                              isActive={subItem.active}
                            >
                              <Link
                                href={{
                                  pathname: subItem.url,
                                  query: {
                                    appId,
                                  },
                                }}
                              >
                                <span>{subItem.name}</span>
                              </Link>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </SidebarMenuItem>
                </Collapsible>
              ) : (
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
              )
            )}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  );
}
