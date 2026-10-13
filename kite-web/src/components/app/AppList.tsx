import { Button } from "@/components/ui/button";

import { useAppsQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import { toast } from "sonner";
import AppListEntry from "./AppListEntry";
import AppCreateDialog from "./AppCreateDialog";
import { useApps, useResponseData } from "@/lib/hooks/api";
import AutoAnimate from "../common/AutoAnimate";
import { loginUrl } from "@/lib/api/client";
import { BotIcon } from "lucide-react";

export default function AppList() {
  const apps = useApps((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load apps: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unauthorized") {
        window.location.href = loginUrl;
      }
    }
  });

  return (
    <AutoAnimate className="flex flex-col space-y-5">
      {!apps ? (
        <>
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </>
      ) : apps.length === 0 ? (
        <div className="rounded-lg border border-dashed p-6 text-center">
          <BotIcon className="h-10 w-10 mx-auto text-muted-foreground" />
          <h2 className="text-xl font-bold mt-3">Add your first bot</h2>
          <p className="text-sm text-muted-foreground mt-2 mb-5">
            Each app in Kite is one Discord bot.
          </p>
          <AppCreateDialog>
            <Button>Create app</Button>
          </AppCreateDialog>
        </div>
      ) : (
        apps.map((app) => <AppListEntry app={app!} key={app?.id} />)
      )}
      {/* Also shown if the apps failed to load, so creating one still works. */}
      {apps?.length !== 0 && (
        <div className="flex">
          <AppCreateDialog>
            <Button>Create app</Button>
          </AppCreateDialog>
        </div>
      )}
    </AutoAnimate>
  );
}
