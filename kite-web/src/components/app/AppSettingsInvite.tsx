import Link from "next/link";
import { Button } from "../ui/button";
import { ExternalLinkIcon } from "lucide-react";
import { useApp } from "@/lib/hooks/api";
import { getAppInviteUrl } from "@/lib/discord/invite";

export default function AppSettingsInvite() {
  const app = useApp();

  if (!app) return null;

  return (
    <Button asChild className="w-full md:w-auto">
      <Link href={getAppInviteUrl(app.discord_id)} target="_blank">
        <ExternalLinkIcon className="h-5 w-5 mr-2.5" />
        Invite bot
      </Link>
    </Button>
  );
}
