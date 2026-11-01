import { Button } from "@/components/ui/button";
import { Skeleton } from "../ui/skeleton";
import { toast } from "sonner";
import AppListEntry from "./AppListEntry";
import AppCreateDialog from "./AppCreateDialog";
import { useApps } from "@/lib/hooks/api";
import AutoAnimate from "../common/AutoAnimate";
import { loginUrl } from "@/lib/api/client";
import { BotIcon, PlusIcon } from "lucide-react";

const firstSteps = [
  "Create a bot in the Discord Developer Portal",
  "Connect it to Kite with its token",
  "Build commands and events in the visual editor",
];

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

  if (apps?.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed bg-card/50 px-6 py-10 text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <BotIcon className="h-7 w-7" />
        </div>
        <h2 className="text-xl font-semibold tracking-tight">
          Create your first bot
        </h2>
        <p className="mx-auto mt-1.5 max-w-sm text-sm text-muted-foreground">
          Each app in Kite is one Discord bot. Setting one up takes a couple of
          minutes.
        </p>

        <ol className="mx-auto mt-6 w-fit space-y-2.5 text-left">
          {firstSteps.map((step, i) => (
            <li
              key={step}
              className="flex items-center gap-3 text-sm text-muted-foreground"
            >
              <span className="flex h-6 w-6 flex-none items-center justify-center rounded-full border bg-background text-xs font-medium text-foreground">
                {i + 1}
              </span>
              {step}
            </li>
          ))}
        </ol>

        <AppCreateDialog>
          <Button className="mt-7 gap-1.5">
            <PlusIcon className="h-4 w-4" />
            Create app
          </Button>
        </AppCreateDialog>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <AutoAnimate className="flex flex-col gap-3">
        {!apps ? (
          <>
            <Skeleton className="h-[86px] rounded-xl" />
            <Skeleton className="h-[86px] rounded-xl" />
            <Skeleton className="h-[86px] rounded-xl" />
          </>
        ) : (
          apps.map((app) => <AppListEntry app={app!} key={app?.id} />)
        )}
      </AutoAnimate>

      {/* Also shown if the apps failed to load, so creating one still works. */}
      <AppCreateDialog>
        <button className="flex w-full items-center justify-center gap-2 rounded-xl border border-dashed p-4 text-sm font-medium text-muted-foreground outline-none transition-colors hover:border-primary/40 hover:bg-accent/40 hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring">
          <PlusIcon className="h-4 w-4" />
          Create app
        </button>
      </AppCreateDialog>
    </div>
  );
}
