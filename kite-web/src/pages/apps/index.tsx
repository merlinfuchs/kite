import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import LegalUpdateDialog from "@/components/common/LegalUpdateDialog";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import AccountPageLayout from "@/components/app/AccountPageLayout";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CircleAlertIcon } from "lucide-react";
import { useRouter } from "next/router";

// Why adding the Kite bot didn't give the server an app, from the service.
const installErrors: Record<string, string> = {
  app_limit:
    "You already have 10 apps, the most an account can have. Delete one, then add the Kite bot again.",
  banned: "Kite banned your account, so you can't add apps.",
};

export default function AppListPage() {
  const router = useRouter();
  const installError = installErrors[String(router.query.install_error)];

  return (
    <BaseLayout title="Apps" noindex>
      <LegalUpdateDialog />
      <AccountPageLayout>
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Your apps</h1>
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
        {installError && (
          <Alert variant="destructive" className="mb-4">
            <CircleAlertIcon className="h-4 w-4" />
            <AlertDescription>{installError}</AlertDescription>
          </Alert>
        )}
        <AppList />
      </AccountPageLayout>
    </BaseLayout>
  );
}
