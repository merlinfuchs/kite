import logo from "@/assets/logo/white@1024.png";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthLogoutMutation } from "@/lib/api/mutations";
import env from "@/lib/env/client";
import { useApp, useUser } from "@/lib/hooks/api";
import { abbreviateName, cn } from "@/lib/utils";
import {
  HomeIcon,
  MailPlusIcon,
  MessageSquareWarningIcon,
  NotebookTabsIcon,
  PanelLeft,
  SatelliteDishIcon,
  ServerIcon,
  SettingsIcon,
  SlashSquareIcon,
  SmilePlusIcon,
  VariableIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { Fragment, ReactNode, useCallback, useMemo } from "react";
import { toast } from "sonner";
import BaseLayout from "../common/BaseLayout";
import ThemeSwitch from "../common/ThemeSwitch";
import DynamicIcon from "../icons/DynamicIcon";
import { Separator } from "../ui/separator";
import AppDisabledPopup from "./AppDisabledPopup";
import OpenBetaPopup from "./OpenBetaPopup";

interface Props {
  breadcrumbs?: {
    label: string;
    href?: string;
  }[];
  title?: string;
  children: ReactNode;
  disablePadding?: boolean;
}

export default function AppLayout({ children, ...props }: Props) {
  const router = useRouter();
  const logoutMutation = useAuthLogoutMutation();

  const logout = useCallback(() => {
    router.push("/");
    setTimeout(() => logoutMutation.mutate(), 500);
  }, [logoutMutation, router]);

  const user = useUser((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load user: ${res.error.message} (${res.error.code})`
      );
      if (res.error.code === "unauthorized") {
        router.push("/login");
      }
    }
  });

  const app = useApp((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load app: ${res.error.message} (${res.error.code})`
      );
      if (res.error.code === "unknown_app") {
        router.push("/apps");
      }
    }
  });

  const breadcrumbs = useMemo(
    () => [
      {
        label: "Apps",
        href: "/apps",
      },
      {
        label: app?.name || "Unknown App",
        href: props.breadcrumbs?.length ? `/apps/[appId]` : undefined,
      },
      ...(props.breadcrumbs || []),
    ],
    [app, props.breadcrumbs]
  );
  const title = useMemo(
    () => props.title || app?.name || "Unknown App",
    [app, props.title]
  );

  /*const appId = useAppId();
  const dashboards = useResponseData(useAppDashboardsQuery(projectId));
  const dashboardPages = useMemo(() => {
    if (!dashboards) return [];

    return dashboards.flatMap((dashboard) => {
      if (!dashboard) return [];

      return dashboard.pages.map((p) => {
        const encodedPath = encodeURIComponent(p.path);

        return {
          appId: dashboard.app_id,
          icon: p.icon,
          label: p.label,
          path: p.path,
          href: `/projects/[pid]/apps/[appId]/dashboard`,
          active: router.asPath.startsWith(
            `/projects/${projectId}/apps/${dashboard.app_id}/dashboard?path=${encodedPath}`
          ),
        };
      });
    });
  }, [dashboards, router.asPath]); */

  const dashboardPages = [] as any[];

  const navItems = useMemo(() => {
    function isActive(path: string, exact = false) {
      if (exact) {
        return router.pathname === path;
      }

      return router.pathname.startsWith(path);
    }

    return [
      {
        icon: HomeIcon,
        label: "Overview",
        href: `/apps/[appId]`,
        active: isActive(`/apps/[appId]`, true),
      },
      {
        icon: SlashSquareIcon,
        label: "Custom Commands",
        href: `/apps/[appId]/commands`,
        active: isActive(`/apps/[appId]/commands`),
      },
      {
        icon: SatelliteDishIcon,
        label: "Event Listeners",
        href: `/apps/[appId]/events`,
        active: isActive(`/apps/[appId]/events`),
      },
      {
        icon: MailPlusIcon,
        label: "Message Templates",
        href: `/apps/[appId]/messages`,
        active: isActive(`/apps/[appId]/messages`),
      },
      {
        icon: VariableIcon,
        label: "Stored Variables",
        href: `/apps/[appId]/variables`,
        active: isActive(`/apps/[appId]/variables`),
      },
      {
        icon: SmilePlusIcon,
        label: "Emoji Explorer",
        href: `/apps/[appId]/emojis`,
        active: isActive(`/apps/[appId]/emojis`),
        bottom: true,
      },
      {
        icon: ServerIcon,
        label: "Server Explorer",
        href: `/apps/[appId]/guilds`,
        active: isActive(`/apps/[appId]/guilds`),
        bottom: true,
      },
      {
        icon: MessageSquareWarningIcon,
        label: "Logs",
        href: `/apps/[appId]/logs`,
        active: isActive(`/apps/[appId]/logs`),
        bottom: true,
      },
      {
        icon: SettingsIcon,
        label: "Settings",
        href: `/apps/[appId]/settings`,
        active: isActive(`/apps/[appId]/settings`),
        bottom: true,
      },
    ];
  }, [router.pathname]);

  return (
    <BaseLayout title={title}>
      <div className="flex min-h-[100dvh] w-full flex-col bg-muted/40">
        <aside className="fixed inset-y-0 left-0 z-10 hidden w-14 flex-col border-r bg-background sm:flex">
          <nav className="flex flex-col items-center gap-4 px-2 sm:py-5">
            <Link
              href="/"
              className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:h-8 md:w-8 md:text-base"
            >
              <img
                src={logo.src}
                className="h-6 w-6 transition-all group-hover:scale-105"
                alt="Kite Logo"
              />
              <span className="sr-only">Kite.onl</span>
            </Link>
            {navItems
              .filter((i) => !i.bottom)
              .map((item) => (
                <Tooltip key={item.href}>
                  <TooltipTrigger asChild>
                    <Link
                      href={{
                        pathname: item.href,
                        query: router.query,
                      }}
                      className={cn(
                        "flex h-9 w-9 items-center justify-center rounded-lg transition-colors hover:text-foreground md:h-8 md:w-8",
                        item.active
                          ? "bg-accent text-accent-foreground"
                          : "text-muted-foreground"
                      )}
                    >
                      <item.icon className="h-5 w-5" />
                      <span className="sr-only">{item.label}</span>
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent side="right">{item.label}</TooltipContent>
                </Tooltip>
              ))}

            {dashboardPages.length > 0 && <Separator className="my-1" />}
            {dashboardPages.map((page) => (
              <Tooltip key={page.href}>
                <TooltipTrigger asChild>
                  <Link
                    href={{
                      pathname: page.href,
                      query: {
                        pid: router.query.pid,
                        appId: page.appId,
                        path: page.path,
                      },
                    }}
                    className={cn(
                      "flex h-9 w-9 items-center justify-center rounded-lg transition-colors hover:text-foreground md:h-8 md:w-8",
                      page.active
                        ? "bg-accent text-accent-foreground"
                        : "text-muted-foreground"
                    )}
                  >
                    <DynamicIcon name={page.icon} className="h-5 w-5" />
                    <span className="sr-only">{page.label}</span>
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right">{page.label}</TooltipContent>
              </Tooltip>
            ))}
          </nav>
          <nav className="mt-auto flex flex-col items-center gap-4 px-2 sm:py-5">
            {navItems
              .filter((i) => i.bottom)
              .map((item) => (
                <Tooltip key={item.href}>
                  <TooltipTrigger asChild>
                    <Link
                      href={{
                        pathname: item.href,
                        query: router.query,
                      }}
                      className={cn(
                        "flex h-9 w-9 items-center justify-center rounded-lg transition-colors hover:text-foreground md:h-8 md:w-8",
                        item.active
                          ? "bg-accent text-accent-foreground"
                          : "text-muted-foreground"
                      )}
                    >
                      <item.icon className="h-5 w-5" />
                      <span className="sr-only">{item.label}</span>
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent side="right">{item.label}</TooltipContent>
                </Tooltip>
              ))}
          </nav>
        </aside>
        <div className="flex flex-1 flex-col sm:gap-6 sm:py-4 sm:pl-14 w-full max-w-[1500px] mx-auto">
          <header className="sticky top-0 z-30 flex h-14 items-center gap-5 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
            <Sheet>
              <SheetTrigger asChild>
                <Button size="icon" variant="outline" className="sm:hidden">
                  <PanelLeft className="h-5 w-5" />
                  <span className="sr-only">Toggle Menu</span>
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="sm:max-w-xs">
                <nav className="grid gap-6 text-lg font-medium">
                  <Link
                    href="/"
                    className="group flex h-10 w-10 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:text-base"
                  >
                    <img
                      src={logo.src}
                      className="h-6 w-6 transition-all group-hover:scale-105"
                      alt="Kite Logo"
                    />
                    <span className="sr-only">Kite.onl</span>
                  </Link>

                  {navItems.map((item) => (
                    <Link
                      key={item.href}
                      href={{
                        pathname: item.href,
                        query: router.query,
                      }}
                      className={cn(
                        "flex items-center gap-4 px-2.5 hover:text-foreground",
                        item.active
                          ? "text-accent-foreground"
                          : "text-muted-foreground"
                      )}
                    >
                      <item.icon className="h-5 w-5" />
                      {item.label}
                    </Link>
                  ))}

                  {dashboardPages.length > 0 && <Separator className="my-1" />}
                  {dashboardPages.map((page) => (
                    <Link
                      key={page.href}
                      href={{
                        pathname: page.href,
                        query: {
                          pid: router.query.pid,
                          appId: page.appId,
                          path: page.path,
                        },
                      }}
                      className={cn(
                        "flex items-center gap-4 px-2.5 hover:text-foreground",
                        page!.active
                          ? "text-accent-foreground"
                          : "text-muted-foreground"
                      )}
                    >
                      <DynamicIcon name={page!.icon} className="h-5 w-5" />
                      {page.label}
                    </Link>
                  ))}
                </nav>
              </SheetContent>
            </Sheet>
            <Breadcrumb className="hidden md:flex">
              <BreadcrumbList>
                {breadcrumbs.map((item, i) => (
                  <Fragment key={item.label}>
                    <BreadcrumbItem>
                      {item.href ? (
                        <BreadcrumbLink asChild>
                          <Link
                            href={{
                              pathname: item.href,
                              query: router.query,
                            }}
                          >
                            {item.label}
                          </Link>
                        </BreadcrumbLink>
                      ) : (
                        <BreadcrumbPage>{item.label}</BreadcrumbPage>
                      )}
                    </BreadcrumbItem>

                    {i < breadcrumbs.length - 1 && <BreadcrumbSeparator />}
                  </Fragment>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
            <div className="relative ml-auto flex-shrink-0">
              <ThemeSwitch />
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  className="overflow-hidden rounded-full"
                >
                  <div className="font-medium tracking-wide">
                    {abbreviateName(user?.display_name || "")}
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>
                  {user?.display_name || ""}
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild className="cursor-pointer">
                  <Link href={env.NEXT_PUBLIC_DOCS_LINK} target="_blank">
                    Docs
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild className="cursor-pointer">
                  <Link href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank">
                    Discord
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout} className="cursor-pointer">
                  Logout
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </header>
          <main
            className={cn(
              "flex-1 flex flex-col",
              !props.disablePadding && "p-4 sm:px-6 sm:pb-20"
            )}
          >
            {children}
          </main>
        </div>
      </div>

      <OpenBetaPopup />
      <AppDisabledPopup />
    </BaseLayout>
  );
}
