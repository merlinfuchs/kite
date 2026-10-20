import Link from "next/link";
import logo from "@/assets/logo/orange@1024.png";
import env from "@/lib/env/client";
import HomeNavbarMenu from "./HomeNavbarMenu";
import HomeNavbarMobile from "./HomeNavbarMobile";
import { Button } from "../ui/button";
import { useResponseData } from "@/lib/hooks/api";
import { useUserQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import ThemeSwitch from "../common/ThemeSwitch";
import DiscordIcon from "../icons/DiscordIcon";
import { loginUrl } from "@/lib/api/client";

export default function HomeNavbar() {
  const userQuery = useUserQuery();
  const user = useResponseData(userQuery);

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center gap-2 px-4 sm:px-8">
        <Link
          href="/"
          className="mr-2 flex items-center gap-2 rounded-md font-semibold text-base outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <img src={logo.src} className="h-6 w-6" alt="" />
          Kite
        </Link>

        <HomeNavbarMenu className="hidden md:flex" />

        <div className="ml-auto flex items-center gap-1 sm:gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="hidden h-9 w-9 text-muted-foreground hover:text-foreground md:inline-flex"
            asChild
          >
            <a
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              rel="noopener"
              aria-label="Join our Discord server"
              title="Join our Discord server"
            >
              <DiscordIcon className="h-[18px] w-[18px]" />
            </a>
          </Button>
          <ThemeSwitch className="m-0 h-9 w-9 text-muted-foreground hover:text-foreground [&_svg]:h-[18px] [&_svg]:w-[18px]" />

          <div className="mx-1 hidden h-5 w-px bg-border sm:block" />

          {userQuery.isPending ? (
            <Skeleton className="h-9 w-20 sm:w-32" />
          ) : user ? (
            <Button size="sm" asChild>
              <Link href="/apps">Open app</Link>
            </Button>
          ) : (
            <Button size="sm" asChild>
              <a href={loginUrl}>Log in</a>
            </Button>
          )}

          <HomeNavbarMobile user={user} className="md:hidden" />
        </div>
      </div>
    </header>
  );
}
