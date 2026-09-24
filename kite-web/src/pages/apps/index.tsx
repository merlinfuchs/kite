import AppList from "@/components/app/AppList";
import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useAuthLogoutMutation } from "@/lib/api/mutations";
import { useUser } from "@/lib/hooks/api";
import { LogOutIcon } from "lucide-react";
import { useRouter } from "next/router";
import { useCallback } from "react";

export default function AppListPage() {
  const router = useRouter();
  const user = useUser();

  const logoutMutation = useAuthLogoutMutation();

  const logout = useCallback(() => {
    router.push("/");
    setTimeout(() => logoutMutation.mutate(), 500);
  }, [logoutMutation, router]);

  return (
    <BaseLayout title="Apps">
      <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
        <div className="w-full max-w-lg">
          <div>
            <h1 className="text-lg font-semibold md:text-2xl mb-1">
              Your Apps
            </h1>
            <p className="text-muted-foreground text-sm">
              Apps are where you manage your plugins, integrations, and
              settings. Create an app or ask your team to invite you.
            </p>
          </div>
          <Separator className="my-4" />
          <AppList />
          <div className="flex items-center justify-between mt-6 text-sm text-muted-foreground">
            <div className="truncate">
              {user ? `Logged in as ${user.display_name}` : null}
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="flex gap-2 flex-none"
              onClick={logout}
            >
              <LogOutIcon className="size-4" />
              Log out
            </Button>
          </div>
        </div>
      </div>
    </BaseLayout>
  );
}
