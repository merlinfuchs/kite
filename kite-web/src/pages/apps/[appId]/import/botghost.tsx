import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import BotGhostImport from "@/components/app/BotGhostImport";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Commands",
    href: "/apps/[appId]/commands",
  },
  {
    label: "Import from BotGhost",
  },
];

export default function AppBotGhostImportPage() {
  return (
    <AppLayout title="Import from BotGhost" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Import from BotGhost
        </h1>
        <p className="text-muted-foreground text-sm">
          Bring the custom commands and custom events of your BotGhost bot to
          Kite, so you don&apos;t have to build them again.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/botghost-import`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn more
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <BotGhostImport />
    </AppLayout>
  );
}

AppBotGhostImportPage.getLayout = getAppShellLayout;
