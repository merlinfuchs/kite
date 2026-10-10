import { Button } from "@/components/ui/button";

import { useAppsQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import { toast } from "sonner";
import AppListEntry from "./AppListEntry";
import AppCreateDialog from "./AppCreateDialog";
import { useApps, useResponseData } from "@/lib/hooks/api";
import AutoAnimate from "../common/AutoAnimate";
import { loginUrl } from "@/lib/api/client";

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
      ) : (
        apps.map((app) => <AppListEntry app={app!} key={app?.id} />)
      )}
      <div className="flex">
        <AppCreateDialog>
          <Button>Create app</Button>
        </AppCreateDialog>
      </div>
    </AutoAnimate>
  );
}
