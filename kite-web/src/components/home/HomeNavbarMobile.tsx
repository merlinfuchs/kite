import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  BookOpenIcon,
  LogInIcon,
  MenuIcon,
  TagIcon,
} from "lucide-react";
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
import { User } from "@/lib/types/wire.gen";
import { loginUrl } from "@/lib/api/client";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import DiscordIcon from "../icons/DiscordIcon";
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
      <DrawerContent className="max-h-[85dvh] rounded-t-2xl after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
        <DrawerHeader className="sr-only">
          <DrawerTitle>Navigation menu</DrawerTitle>
          <DrawerDescription>
            Site navigation options for Kite
          </DrawerDescription>
        </DrawerHeader>

        <nav className="flex-1 overflow-y-auto no-scrollbar px-3 pt-3 pb-3 space-y-4">
          <div>
            <div className="px-1 pb-1.5 text-xs font-medium text-muted-foreground">
              Product
            </div>
            <div className="grid grid-cols-1 min-[300px]:grid-cols-2 gap-1">
              {productLinks.map((link) => {
                const isActive = router.pathname === link.href;
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={close}
                    className={cn(
                      "flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                      isActive && "bg-accent text-accent-foreground"
                    )}
                  >
                    <span className="flex size-7 flex-none items-center justify-center rounded-md bg-primary/10 text-primary">
                      <link.icon className="size-4" />
                    </span>
                    <span className="truncate">{link.title}</span>
                  </Link>
                );
              })}
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between px-1 pb-1.5">
              <div className="text-xs font-medium text-muted-foreground">
                Tools
              </div>
              <Link
                href="/tools"
                onClick={close}
                className={cn(
                  "flex items-center gap-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground",
                  router.pathname === "/tools" && "text-foreground"
                )}
              >
                View all tools
                <ArrowRightIcon className="size-3.5" />
              </Link>
            </div>
            <div className="flex flex-col gap-0.5">
              {navTools.map((tool) => {
                const isActive =
                  !tool.external && router.pathname === tool.href;
                const rowClass = cn(
                  "flex items-center gap-2.5 rounded-lg px-2 py-1.5 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                  isActive && "bg-accent text-accent-foreground"
                );
                const content = (
                  <>
                    <span
                      className={cn(
                        "flex size-7 flex-none items-center justify-center rounded-md",
                        tool.color
                      )}
                    >
                      <tool.icon className="size-4" />
                    </span>
                    <span className="truncate">{tool.title}</span>
                    {tool.external && (
                      <ArrowUpRightIcon className="ml-auto size-3.5 flex-none text-muted-foreground" />
                    )}
                  </>
                );

                return tool.external ? (
                  <a
                    key={tool.id}
                    href={tool.href}
                    target="_blank"
                    rel="noopener"
                    onClick={close}
                    className={rowClass}
                  >
                    {content}
                  </a>
                ) : (
                  <Link
                    key={tool.id}
                    href={tool.href}
                    onClick={close}
                    className={rowClass}
                  >
                    {content}
                  </Link>
                );
              })}
            </div>
          </div>
        </nav>

        <div className="border-t border-border/60 p-3 shrink-0 space-y-2.5">
          <div className="grid grid-cols-3 gap-1.5">
            <Link
              href="/pricing"
              onClick={close}
              className={cn(
                "flex items-center justify-center gap-1 rounded-lg border border-border/60 bg-muted/40 px-2 py-2 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                router.pathname === "/pricing" &&
                  "bg-accent text-accent-foreground"
              )}
            >
              <TagIcon className="size-3.5 text-primary" />
              <span>Pricing</span>
            </Link>
            <a
              href={env.NEXT_PUBLIC_DOCS_LINK}
              target="_blank"
              rel="noopener"
              onClick={close}
              className="flex items-center justify-center gap-1 rounded-lg border border-border/60 bg-muted/40 px-2 py-2 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <BookOpenIcon className="size-3.5 text-muted-foreground" />
              <span>Docs</span>
              <ArrowUpRightIcon className="hidden min-[340px]:block size-3 text-muted-foreground" />
            </a>
            <a
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              rel="noopener"
              onClick={close}
              className="flex items-center justify-center gap-1 rounded-lg border border-border/60 bg-muted/40 px-2 py-2 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <DiscordIcon className="size-3.5 text-muted-foreground" />
              <span>Discord</span>
              <ArrowUpRightIcon className="hidden min-[340px]:block size-3 text-muted-foreground" />
            </a>
          </div>

          {user ? (
            <Button className="w-full" size="sm" asChild>
              <Link href="/apps" onClick={close}>
                Open app
              </Link>
            </Button>
          ) : (
            <Button className="w-full" size="sm" asChild>
              <a href={loginUrl} onClick={close}>
                <LogInIcon className="mr-2 size-4" />
                Log in
              </a>
            </Button>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  );
}
