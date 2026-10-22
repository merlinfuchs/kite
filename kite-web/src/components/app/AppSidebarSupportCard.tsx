import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { useSidebarSupportStore } from "@/lib/state/sidebarSupport";
import { XIcon } from "lucide-react";
import { useEffect, useState } from "react";

export default function AppSidebarSupportCard() {
  const [mounted, setMounted] = useState(false);
  const dismissedUntil = useSidebarSupportStore((s) => s.dismissedUntil);
  const dismiss = useSidebarSupportStore((s) => s.dismiss);

  useEffect(() => {
    setMounted(true);
  }, []);

  const isDismissed =
    typeof dismissedUntil === "number" && Date.now() < dismissedUntil;

  if (!mounted || isDismissed) {
    return null;
  }

  return (
    <div className="rounded-lg border bg-background p-3 space-y-2 group-data-[collapsible=icon]:hidden">
      <div className="flex items-center justify-between gap-2">
        <div className="text-sm font-medium">Need help?</div>
        <Button
          size="icon"
          variant="ghost"
          className="h-6 w-6 -mr-1 -mt-1 text-muted-foreground hover:text-foreground"
          onClick={dismiss}
          title="Dismiss for 3 days"
          aria-label="Dismiss for 3 days"
        >
          <XIcon className="h-3.5 w-3.5" />
        </Button>
      </div>
      <div className="text-xs text-muted-foreground">
        Ask questions and share feedback in our Discord server.
      </div>
      <Button asChild size="sm" className="w-full">
        <a href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank">
          Join the Discord
        </a>
      </Button>
    </div>
  );
}
