import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useUser } from "@/lib/hooks/api";
import { useLogout } from "@/lib/hooks/auth";
import { HomeIcon, LogOutIcon } from "lucide-react";
import Link from "next/link";

export default function AppListPage() {
  const user = useUser();
  const logout = useLogout();

  return (
    <BaseLayout title="Apps" noindex>
      <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
        <div className="w-full max-w-lg">
          <div>
            <h1 className="text-lg font-semibold md:text-2xl mb-1">
              Your Apps
            </h1>
            <p className="text-muted-foreground text-sm">
              Apps are where you manage your plugins, integrations, and
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
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between mt-6 text-sm text-muted-foreground">
            <div className="truncate">
              {user ? `Logged in as ${user.display_name}` : null}
            </div>
            {/* Offsets the ghost buttons' padding so they line up with the text when stacked. */}
            <div className="flex items-center gap-1 flex-none -ml-3 sm:ml-0">
              <Button variant="ghost" size="sm" className="flex gap-2" asChild>
                <Link href="/">
                  <HomeIcon className="size-4" />
                  Return home
                </Link>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="flex gap-2"
                onClick={logout}
              >
                <LogOutIcon className="size-4" />
                Log out
              </Button>
            </div>
          </div>
        </div>
      </div>
    </BaseLayout>
  );
}
