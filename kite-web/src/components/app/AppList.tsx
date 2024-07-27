import { Button } from "@/components/ui/button";

import { useAppsQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import { toast } from "sonner";
import AppListEntry from "./AppListEntry";
import AppCreateDialog from "./AppCreateDialog";
import { useApps, useResponseData } from "@/lib/hooks/api";
import { useRouter } from "next/router";
import AutoAnimate from "../common/AutoAnimate";

export default function AppList() {
  const router = useRouter();

  const apps = useApps((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load apps: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unauthorized") {
        router.push("/login");
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
