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
import { ArrowLeftIcon } from "lucide-react";
import { Fragment, ReactNode, useCallback, useMemo } from "react";
import BaseLayout from "../common/BaseLayout";
import { useApp } from "@/lib/hooks/api";
import Link from "next/link";
import { useRouter } from "next/router";
import ThemeSwitch from "../common/ThemeSwitch";

interface Props {
  breadcrumbs?: {
    label: string;
    href?: string;
  }[];
  title?: string;
  children: ReactNode;
  disablePadding?: boolean;
  showMobileTitle?: boolean;
  showMobileBack?: boolean;
  backHref?: string;
}

export default function AppLayout({ children, ...props }: Props) {
  const app = useApp();

  const router = useRouter();
  const isDashboard = router.pathname === "/apps/[appId]";
  const showMobileTitle = props.showMobileTitle ?? isDashboard;
  const showMobileBack = props.showMobileBack ?? !isDashboard;

  const handleBack = useCallback(() => {
    if (props.backHref) {
      router.push(props.backHref);
      return;
    }
    if (typeof window !== "undefined" && window.history.length > 1) {
      router.back();
    } else if (app?.id) {
      router.push({
        pathname: "/apps/[appId]",
        query: { appId: app.id },
      });
    } else {
      router.push("/apps");
    }
  }, [props.backHref, app, router]);

  const breadcrumbs = useMemo(() => {
    const list: { label: string; href?: string }[] = [
      {
        label: "Apps",
        href: "/apps",
      },
    ];
    if (app?.name) {
      list.push({
        label: app.name,
        href:
          props.breadcrumbs?.length && app.id ? `/apps/${app.id}` : undefined,
      });
    }
    if (props.breadcrumbs?.length) {
      list.push(...props.breadcrumbs);
    }
    return list;
  }, [app, props.breadcrumbs]);

  const title = useMemo(
    () => props.title || app?.name || "Kite",
    [app, props.title]
  );

  return (
    <BaseLayout title={props.title}>
      <header className="flex h-14 md:h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
        <div className="flex items-center gap-2 justify-between px-4 w-full">
          <div className="flex items-center gap-2">
            <SidebarTrigger className="-ml-2 md:-ml-1 hidden md:flex" />
            {showMobileBack && (
              <Button
                variant="ghost"
                size="icon"
                onClick={handleBack}
                className="md:hidden size-8 -ml-2 text-muted-foreground hover:text-foreground"
                aria-label="Go back"
              >
                <ArrowLeftIcon className="size-4" />
              </Button>
            )}
            {showMobileTitle && (
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
                          <Link href={item.href}>{item.label}</Link>
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
          <div className="pr-2">
            <ThemeSwitch className="text-muted-foreground hover:text-foreground" />
          </div>
        </div>
      </header>
      <main className="p-4 pt-2 md:pt-8 sm:pb-20 sm:px-6 w-full">
        {children}
      </main>
    </BaseLayout>
  );
}
