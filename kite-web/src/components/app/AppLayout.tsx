import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { ArrowLeftIcon, Grid2X2Icon } from "lucide-react";
import { Fragment, ReactNode, useMemo } from "react";
import BaseLayout from "../common/BaseLayout";
import { useAppQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import Link from "next/link";
import { useRouter } from "next/router";
import ThemeSwitch from "../common/ThemeSwitch";
import {
  isFlowEditorRoute,
  usePwaInstall,
  usePwaTransitionReady,
} from "../common/PwaExperience";
import { BotLoadingGate } from "./BotLoadingScreen";

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
  const appId =
    typeof router.query.appId === "string" ? router.query.appId : "";
  const appQuery = useAppQuery(appId);
  const app = useResponseData(appQuery);
  const { isInstalled } = usePwaInstall();
  const isDashboard = router.pathname === "/apps/[appId]";
  const isMarketplace = router.pathname.startsWith("/apps/[appId]/marketplace");
  usePwaTransitionReady(
    !appQuery.isPending && !isFlowEditorRoute(router.pathname)
  );

  const breadcrumbs = useMemo(
    () => [
      {
        label: "Apps",
        href: "/apps",
      },
      {
        label: app?.name || "Unknown app",
        href: props.breadcrumbs?.length ? `/apps/[appId]` : undefined,
      },
      ...(props.breadcrumbs || []),
    ],
    [app, props.breadcrumbs]
  );

  const backHref = useMemo(
    () => breadcrumbs.findLast((item) => item.href)?.href,
    [breadcrumbs]
  );

  const title = useMemo(
    () => props.title || app?.name || "Kite",
    [app, props.title]
  );

  if (isInstalled && isFlowEditorRoute(router.pathname)) return children;

  return (
    <BotLoadingGate
      enabled={isInstalled && isDashboard}
      isLoading={appQuery.isPending}
    >
      <BaseLayout title={props.title} noindex>
        <header className="flex h-14 md:h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
          <div className="flex items-center gap-2 justify-between px-4 w-full">
            <div className="flex items-center gap-2">
              <SidebarTrigger className="-ml-2 md:-ml-1 hidden md:flex" />
              {!isDashboard && backHref && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="md:hidden size-8 -ml-2 text-muted-foreground hover:text-foreground"
                  asChild
                >
                  <Link
                    href={{ pathname: backHref, query: router.query }}
                    aria-label="Go back"
                  >
                    <ArrowLeftIcon className="size-4" />
                  </Link>
                </Button>
              )}
              {isDashboard && (
                <span className="font-semibold text-sm md:hidden truncate max-w-[200px]">
                  {title}
                </span>
              )}
              <Separator
                orientation="vertical"
                className="mr-2 h-4 hidden md:block"
              />
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
            </div>
            <div className="flex items-center gap-1 pr-2">
              {isInstalled && isMarketplace && (
                <Button variant="ghost" size="sm" asChild>
                  <Link href="/apps">
                    <Grid2X2Icon className="mr-2 size-4" />
                    Back to apps
                  </Link>
                </Button>
              )}
              <ThemeSwitch className="text-muted-foreground hover:text-foreground" />
            </div>
          </div>
        </header>
        <main className="p-4 pt-2 md:pt-8 sm:pb-20 sm:px-6 w-full">
          {children}
        </main>
      </BaseLayout>
    </BotLoadingGate>
  );
}
