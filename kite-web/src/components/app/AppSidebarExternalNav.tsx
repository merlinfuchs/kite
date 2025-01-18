import { ExternalLinkIcon } from "lucide-react";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import env from "@/lib/env/client";

const items = [
  {
    name: "Documentation",
    url: env.NEXT_PUBLIC_DOCS_LINK,
  },
  {
    name: "Support",
    url: env.NEXT_PUBLIC_DISCORD_LINK,
  },
];

export default function AppSidebarExternalNav() {
  const appId = useAppId();

  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel>Links</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton asChild>
              <Link href={item.url} target="_blank">
                <span>{item.name}</span>
                <ExternalLinkIcon className="ml-auto text-muted-foreground" />
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
