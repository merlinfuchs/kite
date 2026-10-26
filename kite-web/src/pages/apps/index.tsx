import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import LegalUpdateDialog from "@/components/common/LegalUpdateDialog";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useUser } from "@/lib/hooks/api";
import { ChevronsUpDown } from "lucide-react";
import UserMenu, { UserAvatar } from "@/components/app/UserMenu";

export default function AppListPage() {
  const user = useUser();

  return (
    <BaseLayout title="Apps" noindex>
      <LegalUpdateDialog />
      <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
        <div className="w-full max-w-lg">
          <div>
            <h1 className="text-lg font-semibold md:text-2xl mb-1">
              Your Apps
            </h1>
            <p className="text-muted-foreground text-sm">
              Apps are where you build commands, manage integrations and
              settings. Create an app or ask your team to invite you. Need help?{" "}
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
          {/* Offsets the ghost button's padding so it lines up with the list. */}
          <div className="mt-6 -ml-3">
            <UserMenu align="start">
              <Button
                variant="ghost"
                size="sm"
                className="gap-2 text-muted-foreground"
              >
                <UserAvatar className="size-6 text-xs" />
                {user?.display_name}
                <ChevronsUpDown className="size-4" />
              </Button>
            </UserMenu>
          </div>
        </div>
      </div>
    </BaseLayout>
  );
}
