import BaseLayout from "@/components/common/BaseLayout";
import { Separator } from "@/components/ui/separator";
import { loginUrlTo } from "@/lib/api/client";
import { useUser } from "@/lib/hooks/api";
import { useRouter } from "next/router";
import { ReactNode } from "react";
import AdminNav from "./AdminNav";

// AdminPage is a page where admins moderate the marketplace. Its content is
// only rendered for them, others are sent to log in or told they can't see it.
export default function AdminPage({
  title,
  description,
  className,
  children,
}: {
  title: string;
  description: string;
  className?: string;
  children: ReactNode;
}) {
  const router = useRouter();
  const user = useUser((res) => {
    if (!res.success && res.error.code === "unauthorized") {
      window.location.href = loginUrlTo(router.asPath);
    }
  });

  return (
    <BaseLayout title={title} noindex>
      <div className={className ?? "container max-w-5xl py-12"}>
        <AdminNav />
        <h1 className="text-lg font-semibold md:text-2xl mb-1">{title}</h1>
        <p className="text-muted-foreground text-sm">{description}</p>
        <Separator className="my-8" />
        {user &&
          (user.is_admin ? (
            children
          ) : (
            <p className="text-sm text-muted-foreground">
              You don&apos;t have access to this page.
            </p>
          ))}
      </div>
    </BaseLayout>
  );
}
