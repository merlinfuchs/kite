import {
  VariableIcon,
  KeyRoundIcon,
  PlugIcon,
  StoreIcon,
  SlashSquareIcon,
  MailPlusIcon,
  PackageIcon,
} from "lucide-react";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useBundleInstalls } from "@/lib/hooks/api";
import { hasUpdate } from "@/lib/flow/bundleInstalls";
import { useCallback, useMemo } from "react";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import Link from "next/link";
import { eventSourceName, eventSources } from "@/lib/flow/eventTypes";

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
  const updates =
    useBundleInstalls()?.filter((i) => i && hasUpdate(i)).length ?? 0;

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
          ...eventSources.map((source) => {
            const { plural, path, icon } = eventSourceName(source);
            return { name: plural, url: path, icon };
          }),
          {
            name: "Message Templates",
            url: "/apps/[appId]/messages",
            icon: MailPlusIcon,
          },
        ],
      },
      {
        label: "Library",
        items: [
          {
            name: "Marketplace",
            url: "/apps/[appId]/marketplace",
            icon: StoreIcon,
          },
          {
            name: "Your Templates",
            url: "/apps/[appId]/bundles",
            icon: PackageIcon,
            // Updates are never installed on their own, so they're counted
            // here to be noticed.
            badge: updates || undefined,
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
  }, [updates]);

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
                {"badge" in item && item.badge && (
                  <SidebarMenuBadge>{item.badge}</SidebarMenuBadge>
                )}
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  );
}
