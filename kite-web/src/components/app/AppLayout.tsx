import { AppSidebar } from "@/components/app/AppSidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Fragment, ReactNode, useMemo } from "react";
import BaseLayout from "../common/BaseLayout";
import { useApp } from "@/lib/hooks/api";
import Link from "next/link";
import { useRouter } from "next/router";
import ThemeSwitch from "../common/ThemeSwitch";
import { toast } from "sonner";

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
  const app = useApp((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load app: ${res?.error.message} (${res?.error.code})`
      );
      if (
        res.error.code === "unknown_app" ||
        res.error.code === "missing_access"
      ) {
        router.push({
          pathname: "/apps",
        });
      }
    }
  });

  const router = useRouter();

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

  // TODO: remember open state of sidebar on desktop
  return (
    <BaseLayout title={props.title}>
      <SidebarProvider className="bg-muted/30">
        <AppSidebar />
        <SidebarInset className="bg-transparent min-h-[100dvh] max-w-[1500px] mx-auto">
          <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
            <div className="flex items-center gap-2 justify-between px-4 w-full">
              <div className="flex items-center gap-2">
                <SidebarTrigger className="-ml-1" />
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
              <div className="pr-2">
                <ThemeSwitch />
              </div>
            </div>
          </header>
          <main className="p-4 pt-8 sm:pb-20 sm:px-6 w-full">{children}</main>
        </SidebarInset>
      </SidebarProvider>
    </BaseLayout>
  );
}
