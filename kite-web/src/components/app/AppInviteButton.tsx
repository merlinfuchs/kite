import Link from "next/link";
import { Button } from "../ui/button";
import { ExternalLinkIcon } from "lucide-react";
import { useApp } from "@/lib/hooks/api";
import { getAppInviteUrl } from "@/lib/discord/invite";

export default function AppInviteButton() {
  const app = useApp();

  if (!app) return null;

  return (
    <Button size="sm" variant="default" asChild className="h-8 gap-2">
      <Link href={getAppInviteUrl(app.discord_id)} target="_blank">
        <ExternalLinkIcon className="h-3.5 w-3.5" />
        <span className="lg:sr-only xl:not-sr-only xl:whitespace-nowrap">
          Invite app
        </span>
      </Link>
    </Button>
  );
}
