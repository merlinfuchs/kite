import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import LegalUpdateDialog from "@/components/common/LegalUpdateDialog";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import AccountPageLayout from "@/components/app/AccountPageLayout";

export default function AppListPage() {
  return (
    <BaseLayout title="Apps" noindex>
      <LegalUpdateDialog />
      <AccountPageLayout>
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Your Apps</h1>
          <p className="text-muted-foreground text-sm">
            Apps are where you build commands, manage integrations and settings.
            Create an app or ask your team to invite you. Need help?{" "}
            <a
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              className="text-primary hover:underline"
            >
              Join the Discord server
            </a>
            .
          </p>
        </div>
        <Separator className="my-4" />
        <AppList />
      </AccountPageLayout>
    </BaseLayout>
  );
}
