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
  useSidebar,
} from "@/components/ui/sidebar";
import { useCallback, useMemo } from "react";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import Link from "next/link";
import { eventSourceName, eventSources } from "@/lib/flow/eventTypes";

interface NavItem {
  name: string;
  url: string;
  icon: LucideIcon;
  items?: { name: string; url: string }[];
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

  // Items with sub items are also active on the pages of their sub items.
  const isItemActive = useCallback(
    (item: NavItem) =>
      isActive(item.url) || !!item.items?.some((sub) => isActive(sub.url)),
    [isActive]
  );

  // Sub items are hidden while the sidebar is collapsed to icons, so items
  // with sub items link to their first page then.
  const { state, isMobile } = useSidebar();
  const iconsOnly = state === "collapsed" && !isMobile;

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
            items: eventSources.map((source) => {
              const { plural, path } = eventSourceName(source);
              return { name: plural, url: path };
            }),
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
  }, []);

  return (
    <>
      {sections.map((section) => (
        <SidebarGroup key={section.label}>
          <SidebarGroupLabel>{section.label}</SidebarGroupLabel>
          <SidebarMenu>
            {section.items.map((item) =>
              item.items && !iconsOnly ? (
                <Collapsible
                  key={item.name}
                  asChild
                  defaultOpen={isItemActive(item)}
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
                              isActive={isActive(subItem.url, true)}
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
                  <SidebarMenuButton asChild isActive={isItemActive(item)}>
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
