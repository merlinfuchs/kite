import * as React from "react";
import Link from "next/link";
import logo from "@/assets/logo/orange@1024.png";
import env from "@/lib/env/client";

import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
} from "@/components/ui/navigation-menu";
import { PlusCircleIcon } from "lucide-react";
import { externalTools, kiteTools } from "@/tools/registry";

// Kite's own tools and the message creator. All tools are listed on /tools.
const tools = [
  ...kiteTools,
  externalTools.find((t) => t.id === "message-creator")!,
];

export default function HomeNavbarMenu() {
  return (
    <NavigationMenu>
      <NavigationMenuList>
        <NavigationMenuItem>
          <NavigationMenuTrigger>
            <img
              src={logo.src}
              className="h-6 w-6 mr-2 hidden sm:block"
              alt=""
            />
            Kite
          </NavigationMenuTrigger>
          <NavigationMenuContent>
            <ul className="grid gap-3 p-4 w-[80dvw] md:w-[400px] lg:w-[500px] lg:grid-cols-[.75fr_1fr]">
              <li className="row-span-4">
                <Link href="/" legacyBehavior passHref>
                  <NavigationMenuLink asChild>
                    <div
                      role="button"
                      className="flex h-full w-full select-none flex-col justify-end rounded-md bg-gradient-to-b from-muted/50 to-muted p-6 no-underline outline-none focus:shadow-md"
                    >
                      <img src={logo.src} className="h-10 w-10" alt="" />
                      <div className="my-2 text-lg font-medium">Kite</div>
                      <p className="text-sm leading-tight text-muted-foreground">
                        Custom Discord bots made easy.
                      </p>
                    </div>
                  </NavigationMenuLink>
                </Link>
              </li>
              <ListItem href="/#features" title="Features">
                Learn about the features of Kite and how to use them.
              </ListItem>
              <ListItem href="/#how-it-works" title="How It Works">
                See how you make a Discord bot without writing code.
              </ListItem>
              <ListItem href="/#pricing" title="Pricing">
                Start for free and upgrade a bot when it needs more.
              </ListItem>
              <ListItem href="/#faq" title="Frequently Asked Questions">
                Get answers to common questions about Kite.
              </ListItem>
              <ListItem
                href={env.NEXT_PUBLIC_DOCS_LINK}
                target="_blank"
                title="Documentation"
                className="sm:hidden"
              >
                Learn how to build your bot with Kite.
              </ListItem>
            </ul>
          </NavigationMenuContent>
        </NavigationMenuItem>
        <NavigationMenuItem>
          <NavigationMenuTrigger>Tools</NavigationMenuTrigger>
          <NavigationMenuContent>
            <ul className="grid gap-3 p-4 w-[80dvw] md:w-[500px] md:grid-cols-2 lg:w-[600px]">
              {tools.map((tool) => (
                <ListItem
                  key={tool.title}
                  title={tool.title}
                  href={tool.href}
                  target={tool.external ? "_blank" : undefined}
                >
                  {tool.description}
                </ListItem>
              ))}
              <li>
                <Link href="/tools" legacyBehavior passHref>
                  <NavigationMenuLink asChild>
                    <div className="cursor-pointer flex items-center space-x-4 select-none rounded-md p-3 leading-none no-underline outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground">
                      <PlusCircleIcon className="h-8 w-8 inline-block flex-none text-muted-foreground" />
                      <div className="space-y-1">
                        <div className="text-sm font-medium leading-none">
                          View More
                        </div>
                        <p className="line-clamp-2 text-sm leading-snug text-muted-foreground">
                          View all the tools available on Kite, including the
                          ones not listed here.
                        </p>
                      </div>
                    </div>
                  </NavigationMenuLink>
                </Link>
              </li>
            </ul>
          </NavigationMenuContent>
        </NavigationMenuItem>
        <NavigationMenuItem className="hidden sm:block">
          <NavigationMenuLink
            href={env.NEXT_PUBLIC_DOCS_LINK}
            target="_blank"
            className={navigationMenuTriggerStyle()}
          >
            Documentation
          </NavigationMenuLink>
        </NavigationMenuItem>
        <NavigationMenuItem className="hidden sm:block">
          <NavigationMenuLink
            href={env.NEXT_PUBLIC_DISCORD_LINK}
            target="_blank"
            className={navigationMenuTriggerStyle()}
          >
            Discord
          </NavigationMenuLink>
        </NavigationMenuItem>
      </NavigationMenuList>
    </NavigationMenu>
  );
}

const ListItem = ({
  className,
  title,
  children,
  ...props
}: React.ComponentPropsWithoutRef<typeof Link>) => {
  return (
    <li className={className}>
      <Link {...props} passHref>
        <NavigationMenuLink asChild>
          <div className="cursor-pointer block select-none space-y-1 rounded-md p-3 leading-none no-underline outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground">
            <div className="text-sm font-medium leading-none">{title}</div>
            <p className="line-clamp-2 text-sm leading-snug text-muted-foreground">
              {children}
            </p>
          </div>
        </NavigationMenuLink>
      </Link>
    </li>
  );
};
ListItem.displayName = "ListItem";
