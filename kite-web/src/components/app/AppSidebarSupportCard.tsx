import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";

export default function AppSidebarSupportCard() {
  return (
    <div className="rounded-lg border bg-background p-3 space-y-2 group-data-[collapsible=icon]:hidden">
      <div className="text-sm font-medium">Need help?</div>
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
