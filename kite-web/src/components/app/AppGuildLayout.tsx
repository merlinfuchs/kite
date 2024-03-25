import { ReactNode, useMemo } from "react";
import AppLayout from "./AppLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import Link from "next/link";
import {
  ArchiveBoxIcon,
  Bars3Icon,
  BellIcon,
  CircleStackIcon,
  CodeBracketSquareIcon,
  DocumentArrowUpIcon,
  HomeIcon,
  MagnifyingGlassCircleIcon,
  ShoppingCartIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import { Sheet, SheetContent, SheetTrigger } from "../ui/sheet";
import { Input } from "../ui/input";
import { useUserQuery } from "@/lib/api/queries";
import { useRouter } from "next/router";
import clsx from "clsx";
import AppQuickAccess from "./AppQuickAccess";
import AppGuildSelect from "./AppGuildSelect";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { getApiUrl } from "@/lib/api/client";

/*
<AppTopbar setSidebarOpen={setSidebarOpen} />
      <div className="flex">
        <AppSidebar open={sidebarOpen} setOpen={setSidebarOpen} />
        <div className="pb-32 pt-10 xl:pt-20 px-5 md:px-10 flex flex-auto w-full justify-center">
          <div className="max-w-5xl w-full">{children}</div>
        </div>
      </div>
      */

export default function AppGuildLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const guildId = router.query.gid as string;

  const { data: userResp } = useUserQuery();

  const user = userResp?.success
    ? userResp.data
    : {
        id: "0",
        username: "user",
        global_name: "User",
        discriminator: "0",
        avatar: null,
      };

  const navigation = useMemo(() => {
    return [
      {
        name: "Home",
        href: `/app/guilds/${guildId}`,
        icon: HomeIcon,
        current: router.pathname === "/app/guilds/[gid]",
      },
      {
        name: "Deployments",
        href: `/app/guilds/${guildId}/deployments`,
        icon: DocumentArrowUpIcon,
        current: router.pathname.startsWith(`/app/guilds/[gid]/deployments`),
      },
      {
        name: "Workspaces",
        href: `/app/guilds/${guildId}/workspaces`,
        icon: CodeBracketSquareIcon,
        current: router.pathname.startsWith(`/app/guilds/[gid]/workspaces`),
      },
      {
        name: "KV Storage",
        href: `/app/guilds/${guildId}/kv-storage`,
        icon: CircleStackIcon,
        current: router.pathname.startsWith(`/app/guilds/[gid]/kv-storage`),
      },
      {
        name: "Marketplace",
        href: `/app/guilds/${guildId}/marketplace`,
        icon: ShoppingCartIcon,
        current: router.pathname.startsWith(`/app/guilds/[gid]/marketplace`),
      },
    ];
  }, [guildId]);

  return (
    <AppLayout>
      <div className="grid h-[100dvh] w-full md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
        <div className="hidden border-r bg-muted/40 md:block">
          <div className="flex h-full max-h-screen flex-col gap-2">
            <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
              <Link href="/" className="flex items-center gap-2 font-semibold">
                <img src="/wordmark/white@scalable.svg" alt="Kite Wordmark" />
              </Link>
              {/*<Button variant="outline" size="icon" className="ml-auto h-8 w-8">
                <BellIcon className="h-4 w-4" />
                <span className="sr-only">Toggle notifications</span>
                </Button>*/}
            </div>
            <div className="flex-1 space-y-5">
              <div className="px-2 lg:px-4 mb-7 mt-1">
                <AppGuildSelect guildId={guildId} />
              </div>
              <nav className="grid items-start px-2 text-sm font-medium lg:px-4 space-y-1">
                {navigation.map((item) => (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={clsx(
                      "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary",
                      item.current
                        ? "text-primary bg-muted"
                        : "text-muted-foreground"
                    )}
                  >
                    <item.icon className="h-4 w-4" />
                    {item.name}
                  </Link>
                ))}
              </nav>
              <AppQuickAccess guildId={guildId} />
            </div>
            <div className="mt-auto p-4">
              <Card>
                <CardHeader className="p-2 pt-0 md:p-4">
                  <CardTitle>Need help?</CardTitle>
                  <CardDescription>
                    Join our Discord server for support and information about
                    Kite.
                  </CardDescription>
                </CardHeader>
                <CardContent className="p-2 pt-0 md:p-4 md:pt-0">
                  <Button size="sm" className="w-full">
                    Join Discord
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
        <div className="flex flex-col h-full overflow-y-hidden">
          <header className="flex h-14 items-center gap-4 border-b bg-muted/40 px-4 lg:h-[60px] lg:px-6 flex-none">
            <Sheet>
              <SheetTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  className="shrink-0 md:hidden"
                >
                  <Bars3Icon className="h-5 w-5" />
                  <span className="sr-only">Toggle navigation menu</span>
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="flex flex-col">
                <div className="mb-1 mt-5 mx-[-0.65rem]">
                  <AppGuildSelect guildId={guildId} />
                </div>
                <nav className="grid gap-2 space-y-1 font-medium">
                  {navigation.map((item) => (
                    <Link
                      key={item.name}
                      href={item.href}
                      className={clsx(
                        "mx-[-0.65rem] flex items-center gap-4 rounded-xl px-3 py-2 hover:text-foreground",
                        item.current
                          ? "text-foreground bg-muted"
                          : "text-muted-foreground"
                      )}
                    >
                      <item.icon className="h-6 w-6" />
                      {item.name}
                    </Link>
                  ))}
                </nav>
                <div className="mt-auto">
                  <Card>
                    <CardHeader>
                      <CardTitle>Need help?</CardTitle>
                      <CardDescription>
                        Join our Discord server for support and information
                        about Kite.
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Button size="sm" className="w-full">
                        Join Discord
                      </Button>
                    </CardContent>
                  </Card>
                </div>
              </SheetContent>
            </Sheet>
            <div className="w-full flex-1">
              <form>
                <div className="relative">
                  <MagnifyingGlassCircleIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    type="search"
                    placeholder="Search server..."
                    className="w-full appearance-none bg-background pl-8 shadow-none md:w-2/3 lg:w-1/3"
                  />
                </div>
              </form>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="secondary"
                  size="icon"
                  className="rounded-full"
                >
                  <img
                    src={userAvatarUrl(user)}
                    alt=""
                    className="rounded-full"
                  />
                  <span className="sr-only">Toggle user menu</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>{user.global_name}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <Link href={getApiUrl("/v1/auth/logout")}>Logout</Link>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </header>
          <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-6 overflow-y-auto">
            {children}
          </main>
        </div>
      </div>
    </AppLayout>
  );
}
