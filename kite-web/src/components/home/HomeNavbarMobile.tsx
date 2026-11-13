import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import {
  ExternalLinkIcon,
  LogInIcon,
  MenuIcon,
  TagIcon,
  WrenchIcon,
} from "lucide-react";
import logo from "@/assets/logo/orange@1024.png";
import env from "@/lib/env/client";
import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "../ui/drawer";
import { Separator } from "../ui/separator";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "../ui/sidebar";
import { User } from "@/lib/types/wire.gen";
import { loginUrl } from "@/lib/api/client";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import { navTools, productLinks } from "./navLinks";

export default function HomeNavbarMobile({
  user,
  className,
}: {
  user?: User;
  className?: string;
}) {
  const router = useRouter();
  const isMobile = useIsMobile();
  const [open, setOpen] = useState(false);

  // Close the mobile drawer after navigating or when resizing to desktop.
  useEffect(() => {
    setOpen(false);
  }, [router.asPath, isMobile]);

  const close = () => setOpen(false);

  const externalLinks = [
    { name: "Documentation", url: env.NEXT_PUBLIC_DOCS_LINK },
    { name: "Discord server", url: env.NEXT_PUBLIC_DISCORD_LINK },
    { name: "GitHub repository", url: env.NEXT_PUBLIC_GITHUB_LINK },
  ];

  return (
    <Drawer
      open={open && isMobile}
      onOpenChange={setOpen}
      shouldScaleBackground={false}
    >
      <DrawerTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn("h-9 w-9", className)}
          aria-label="Open menu"
        >
          <MenuIcon className="h-5 w-5" />
        </Button>
      </DrawerTrigger>
      <DrawerContent
        className={cn(
          "fixed inset-x-0 bottom-0 z-50 flex max-h-[85vh] flex-col rounded-t-[16px] border border-b-0 bg-sidebar text-sidebar-foreground shadow-2xl focus:outline-none",
          "after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-sidebar"
        )}
      >
        <DrawerHeader className="sr-only">
          <DrawerTitle>Navigation menu</DrawerTitle>
          <DrawerDescription>
            Site navigation options for Kite
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex items-center gap-2 px-4 pt-3 pb-2 font-semibold text-base shrink-0">
          <img src={logo.src} className="h-6 w-6" alt="" />
          Kite
        </div>

        <Separator className="my-1 bg-sidebar-border" />

        <div className="flex-1 overflow-y-auto no-scrollbar px-2 space-y-2 pb-2 pt-1">
          <SidebarGroup>
            <SidebarGroupLabel>Product</SidebarGroupLabel>
            <SidebarMenu>
              {productLinks.map((link) => (
                <SidebarMenuItem key={link.href}>
                  <SidebarMenuButton
                    asChild
                    isActive={router.pathname === link.href}
                  >
                    <Link href={link.href} onClick={close}>
                      <link.icon />
                      <span>{link.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={router.pathname === "/pricing"}
                >
                  <Link href="/pricing" onClick={close}>
                    <TagIcon />
                    <span>Pricing</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupLabel>Tools</SidebarGroupLabel>
            <SidebarMenu>
              {navTools.map((tool) => (
                <SidebarMenuItem key={tool.id}>
                  <SidebarMenuButton
                    asChild
                    isActive={!tool.external && router.pathname === tool.href}
                  >
                    {tool.external ? (
                      <a
                        href={tool.href}
                        target="_blank"
                        rel="noopener"
                        onClick={close}
                      >
                        <tool.icon />
                        <span>{tool.title}</span>
                        <ExternalLinkIcon className="ml-auto text-muted-foreground" />
                      </a>
                    ) : (
                      <Link href={tool.href} onClick={close}>
                        <tool.icon />
                        <span>{tool.title}</span>
                      </Link>
                    )}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={router.pathname === "/tools"}
                >
                  <Link href="/tools" onClick={close}>
                    <WrenchIcon />
                    <span>All tools</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupLabel>Links</SidebarGroupLabel>
            <SidebarMenu>
              {externalLinks.map((item) => (
                <SidebarMenuItem key={item.name}>
                  <SidebarMenuButton asChild>
                    <a
                      href={item.url}
                      target="_blank"
                      rel="noopener"
                      onClick={close}
                    >
                      <span>{item.name}</span>
                      <ExternalLinkIcon className="ml-auto text-muted-foreground" />
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        </div>

        <div className="border-t border-sidebar-border bg-sidebar p-3 shrink-0">
          {user ? (
            <Button className="w-full" size="sm" asChild>
              <Link href="/apps" onClick={close}>
                Open app
              </Link>
            </Button>
          ) : (
            <Button className="w-full" size="sm" asChild>
              <a href={loginUrl} onClick={close}>
                <LogInIcon className="mr-2 h-4 w-4" />
                Log in
              </a>
            </Button>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  );
}
