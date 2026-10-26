import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import LegalUpdateDialog from "@/components/common/LegalUpdateDialog";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useUser } from "@/lib/hooks/api";
import { ChevronDownIcon } from "lucide-react";
import Link from "next/link";
import logo from "@/assets/logo/orange@1024.png";
import UserMenu, { UserAvatar } from "@/components/app/UserMenu";

export default function AppListPage() {
  const user = useUser();

  return (
    <BaseLayout title="Apps" noindex>
      <LegalUpdateDialog />
      <div className="flex min-h-[100dvh] flex-col">
        <header className="flex h-14 flex-none items-center justify-between border-b px-5">
          <Link
            href="/"
            className="flex items-center gap-2 rounded-md font-semibold outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <img src={logo.src} className="h-6 w-6" alt="" />
            Kite
          </Link>
          <UserMenu returnHome={false}>
            <Button
              variant="ghost"
              size="sm"
              className="gap-2"
              aria-label="Account menu"
            >
              <UserAvatar className="size-6 text-xs" />
              <span className="hidden sm:inline">{user?.display_name}</span>
              <ChevronDownIcon className="size-4 text-muted-foreground" />
            </Button>
          </UserMenu>
        </header>
        <div className="flex flex-1 justify-center items-center w-full px-5 pt-10 pb-20">
          <div className="w-full max-w-lg">
            <div>
              <h1 className="text-lg font-semibold md:text-2xl mb-1">
                Your Apps
              </h1>
              <p className="text-muted-foreground text-sm">
                Apps are where you build commands, manage integrations and
                settings. Create an app or ask your team to invite you. Need
                help?{" "}
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
          </div>
        </div>
      </div>
    </BaseLayout>
  );
}
