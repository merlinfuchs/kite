import * as React from "react";
import Link from "next/link";
import env from "@/lib/env/client";
import { cn } from "@/lib/utils";
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
} from "@/components/ui/navigation-menu";
import { ArrowRightIcon, ArrowUpRightIcon, LucideIcon } from "lucide-react";
import { navTools, productLinks } from "./navLinks";

const triggerClass = "h-9 px-3 bg-transparent";

/**
 * Radix opens a menu on hover after a short delay, and a click on the trigger
 * toggles it. Moving onto a trigger and clicking it therefore opened the menu
 * and immediately closed it again. This keeps a hover-opened menu open when
 * its trigger is clicked (pinning it); a second click closes it as usual.
 */
function useHoverSafeMenu() {
  const [value, setValue] = React.useState("");
  const openedBy = React.useRef<"hover" | "click">("hover");
  const clicking = React.useRef(false);

  const onValueChange = React.useCallback((next: string) => {
    if (next) {
      openedBy.current = clicking.current ? "click" : "hover";
    }
    clicking.current = false;
    setValue(next);
  }, []);

  const triggerProps = React.useCallback(
    (item: string) => ({
      onClick: (e: React.MouseEvent) => {
        if (value === item && openedBy.current === "hover") {
          // Keep it open; the next click will close it.
          e.preventDefault();
          openedBy.current = "click";
          return;
        }
        clicking.current = true;
      },
    }),
    [value]
  );

  return { value, onValueChange, triggerProps };
}

export default function HomeNavbarMenu({ className }: { className?: string }) {
  const { value, onValueChange, triggerProps } = useHoverSafeMenu();

  return (
    <NavigationMenu
      value={value}
      onValueChange={onValueChange}
      className={className}
    >
      <NavigationMenuList>
        <NavigationMenuItem value="product">
          <NavigationMenuTrigger
            className={triggerClass}
            {...triggerProps("product")}
          >
            Product
          </NavigationMenuTrigger>
          <NavigationMenuContent>
            <ul className="grid gap-1 p-2 w-[520px] grid-cols-2">
              {productLinks.map((link) => (
                <MenuItem
                  key={link.href}
                  href={link.href}
                  title={link.title}
                  icon={link.icon}
                  iconClass="bg-primary/10 text-primary"
                >
                  {link.description}
                </MenuItem>
              ))}
            </ul>
          </NavigationMenuContent>
        </NavigationMenuItem>

        <NavigationMenuItem value="tools">
          <NavigationMenuTrigger
            className={triggerClass}
            {...triggerProps("tools")}
          >
            Tools
          </NavigationMenuTrigger>
          <NavigationMenuContent>
            <ul className="grid gap-1 p-2 w-[560px] grid-cols-2">
              {navTools.map((tool) => (
                <MenuItem
                  key={tool.id}
                  href={tool.href}
                  title={tool.title}
                  icon={tool.icon}
                  iconClass={tool.color}
                  external={tool.external}
                >
                  {tool.description}
                </MenuItem>
              ))}
              <li>
                <NavigationMenuLink asChild>
                  <Link
                    href="/tools"
                    className="group/more flex h-full items-center gap-2 rounded-md p-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground outline-none"
                  >
                    View all tools
                    <ArrowRightIcon className="h-4 w-4 transition-transform group-hover/more:translate-x-0.5" />
                  </Link>
                </NavigationMenuLink>
              </li>
            </ul>
          </NavigationMenuContent>
        </NavigationMenuItem>

        <NavigationMenuItem>
          <NavigationMenuLink asChild>
            <Link
              href="/#pricing"
              className={cn(navigationMenuTriggerStyle(), triggerClass)}
            >
              Pricing
            </Link>
          </NavigationMenuLink>
        </NavigationMenuItem>

        <NavigationMenuItem>
          <NavigationMenuLink asChild>
            <a
              href={env.NEXT_PUBLIC_DOCS_LINK}
              target="_blank"
              rel="noopener"
              className={cn(
                navigationMenuTriggerStyle(),
                triggerClass,
                "group/ext"
              )}
            >
              Docs
              <ArrowUpRightIcon className="ml-1 h-3.5 w-3.5 text-muted-foreground transition-transform group-hover/ext:-translate-y-0.5 group-hover/ext:translate-x-0.5" />
            </a>
          </NavigationMenuLink>
        </NavigationMenuItem>
      </NavigationMenuList>
    </NavigationMenu>
  );
}

function MenuItem({
  href,
  title,
  icon: Icon,
  iconClass,
  external,
  children,
}: {
  href: string;
  title: string;
  icon: LucideIcon;
  iconClass: string;
  external?: boolean;
  children: React.ReactNode;
}) {
  const className =
    "flex items-start gap-3 rounded-md p-3 outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground";

  const content = (
    <>
      <div
        className={cn(
          "flex h-9 w-9 flex-none items-center justify-center rounded-md",
          iconClass
        )}
      >
        <Icon className="h-[18px] w-[18px]" />
      </div>
      <div className="space-y-1">
        <div className="flex items-center gap-1 text-sm font-medium leading-none">
          {title}
          {external && (
            <ArrowUpRightIcon className="h-3 w-3 text-muted-foreground" />
          )}
        </div>
        <p className="line-clamp-2 text-xs leading-snug text-muted-foreground">
          {children}
        </p>
      </div>
    </>
  );

  return (
    <li>
      <NavigationMenuLink asChild>
        {external ? (
          <a href={href} target="_blank" rel="noopener" className={className}>
            {content}
          </a>
        ) : (
          <Link href={href} className={className}>
            {content}
          </Link>
        )}
      </NavigationMenuLink>
    </li>
  );
}
