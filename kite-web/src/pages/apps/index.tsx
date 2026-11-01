import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import LegalUpdateDialog from "@/components/common/LegalUpdateDialog";
import DiscordIcon from "@/components/icons/DiscordIcon";
import { Skeleton } from "@/components/ui/skeleton";
import env from "@/lib/env/client";
import { useUser } from "@/lib/hooks/api";
import AccountPageLayout from "@/components/app/AccountPageLayout";
import { BookOpenIcon } from "lucide-react";

export default function AppListPage() {
  const user = useUser();

  return (
    <BaseLayout title="Apps" noindex>
      <LegalUpdateDialog />
      <AccountPageLayout>
        <div className="mb-6">
          {user ? (
            <p className="mb-1 text-sm font-medium text-primary">
              Welcome back, {user.display_name}
            </p>
          ) : (
            <Skeleton className="mb-1.5 h-4 w-40" />
          )}
          <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">
            Your apps
          </h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Apps are where you build commands, manage integrations and settings.
            Create an app or ask your team to invite you.
          </p>
        </div>

        <AppList />

        <div className="mt-8 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-sm text-muted-foreground">
          <span>Need help?</span>
          <a
            href={env.NEXT_PUBLIC_DISCORD_LINK}
            target="_blank"
            rel="noopener"
            className="inline-flex items-center gap-1.5 hover:text-foreground"
          >
            <DiscordIcon className="h-4 w-4" />
            Discord server
          </a>
          <a
            href={env.NEXT_PUBLIC_DOCS_LINK}
            target="_blank"
            rel="noopener"
            className="inline-flex items-center gap-1.5 hover:text-foreground"
          >
            <BookOpenIcon className="h-4 w-4" />
            Docs
          </a>
        </div>
      </AccountPageLayout>
    </BaseLayout>
  );
}
