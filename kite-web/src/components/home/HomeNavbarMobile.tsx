import Link from "next/link";
import { ArrowUpRightIcon, LogInIcon, MenuIcon, TagIcon } from "lucide-react";
import logo from "@/assets/logo/orange@1024.png";
import env from "@/lib/env/client";
import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "../ui/sheet";
import { User } from "@/lib/types/wire.gen";
import { loginUrl } from "@/lib/api/client";
import { navTools, productLinks } from "./navLinks";
import DiscordIcon from "../icons/DiscordIcon";

const rowClass =
  "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground";

export default function HomeNavbarMobile({
  user,
  className,
}: {
  user?: User;
  className?: string;
}) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn("h-9 w-9", className)}
          aria-label="Open menu"
        >
          <MenuIcon className="h-5 w-5" />
        </Button>
      </SheetTrigger>
      <SheetContent
        side="right"
        className="flex w-[85vw] max-w-sm flex-col gap-0 p-0"
      >
        <SheetHeader className="border-b px-5 py-4 text-left">
          <SheetTitle className="flex items-center gap-2">
            <img src={logo.src} className="h-6 w-6" alt="" />
            Kite
          </SheetTitle>
          <SheetDescription className="sr-only">
            Site navigation
          </SheetDescription>
        </SheetHeader>

        <nav className="flex-auto overflow-y-auto px-2 py-4 space-y-6">
          <Section title="Product">
            {productLinks.map((link) => (
              <SheetClose asChild key={link.href}>
                <Link href={link.href} className={rowClass}>
                  <link.icon className="h-4 w-4 text-primary" />
                  {link.title}
                </Link>
              </SheetClose>
            ))}
            <SheetClose asChild>
              <Link href="/#pricing" className={rowClass}>
                <TagIcon className="h-4 w-4 text-primary" />
                Pricing
              </Link>
            </SheetClose>
          </Section>

          <Section title="Tools">
            {navTools.map((tool) => (
              <SheetClose asChild key={tool.id}>
                <a
                  href={tool.href}
                  target={tool.external ? "_blank" : undefined}
                  rel={tool.external ? "noopener" : undefined}
                  className={rowClass}
                >
                  <span
                    className={cn(
                      "flex h-6 w-6 items-center justify-center rounded",
                      tool.color
                    )}
                  >
                    <tool.icon className="h-3.5 w-3.5" />
                  </span>
                  {tool.title}
                  {tool.external && (
                    <ArrowUpRightIcon className="h-3.5 w-3.5 text-muted-foreground" />
                  )}
                </a>
              </SheetClose>
            ))}
            <SheetClose asChild>
              <Link
                href="/tools"
                className={cn(rowClass, "text-muted-foreground")}
              >
                View all tools
              </Link>
            </SheetClose>
          </Section>

          <Section title="Resources">
            <a
              href={env.NEXT_PUBLIC_DOCS_LINK}
              target="_blank"
              rel="noopener"
              className={rowClass}
            >
              Documentation
              <ArrowUpRightIcon className="h-3.5 w-3.5 text-muted-foreground" />
            </a>
            <a
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              rel="noopener"
              className={rowClass}
            >
              <DiscordIcon className="h-4 w-4" />
              Discord server
              <ArrowUpRightIcon className="h-3.5 w-3.5 text-muted-foreground" />
            </a>
          </Section>
        </nav>

        <div className="border-t p-4">
          {user ? (
            <Button className="w-full" asChild>
              <Link href="/apps">Open app</Link>
            </Button>
          ) : (
            <Button className="w-full" asChild>
              <a href={loginUrl}>
                <LogInIcon className="mr-2 h-4 w-4" />
                Log in
              </a>
            </Button>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="px-3 pb-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {title}
      </div>
      <div className="flex flex-col">{children}</div>
    </div>
  );
}
