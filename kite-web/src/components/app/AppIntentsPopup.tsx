import { useApp } from "@/lib/hooks/api";
import { useAppIntentsQuery } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import { TriangleAlertIcon, XIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "../ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";

const dismissKey = (appId: string) => `kite-intents-popup-dismissed:${appId}`;

export default function AppIntentsPopup() {
  const app = useApp();
  const appId = useAppId();

  // The disabled popup takes the same spot and matters more, and a disabled
  // app has no gateway connection that intents would apply to.
  const appUsable = !!app && app.enabled;

  const query = useAppIntentsQuery(appId, appUsable);
  const intents = query.data?.success ? query.data.data : undefined;

  const [dismissed, setDismissed] = useState(true);

  useEffect(() => {
    if (!appId) return;
    try {
      setDismissed(sessionStorage.getItem(dismissKey(appId)) === "1");
    } catch {
      setDismissed(false);
    }
  }, [appId]);

  function dismiss() {
    setDismissed(true);
    try {
      sessionStorage.setItem(dismissKey(appId), "1");
    } catch {
      // Storage can be unavailable, the popup just comes back next load.
    }
  }

  if (!appUsable || !intents || dismissed) return null;

  const missing: string[] = [];
  if (!intents.message_content) missing.push("Message Content Intent");
  if (!intents.guild_members) missing.push("Server Members Intent");

  if (missing.length === 0) return null;

  const portalUrl = `https://discord.com/developers/applications/${app.discord_id}/bot`;

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-yellow-500">
      <CardHeader className="px-5 py-4">
        <CardTitle className="text-base flex items-center gap-2 pr-6">
          <TriangleAlertIcon className="w-5 h-5 text-yellow-500 shrink-0" />
          Privileged Intents Disabled
        </CardTitle>
        <CardDescription>
          Your bot doesn&apos;t have{" "}
          <span className="text-foreground">{missing.join(" and ")}</span>{" "}
          enabled. Some event listeners and blocks won&apos;t work until you
          turn {missing.length > 1 ? "them" : "it"} on under{" "}
          <span className="text-foreground">Privileged Gateway Intents</span> in
          the Discord developer portal.
        </CardDescription>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-0 right-1"
          onClick={dismiss}
        >
          <XIcon className="w-5 h-5 cursor-pointer" />
        </Button>
      </CardHeader>
      <CardFooter className="px-5 pb-4 flex gap-3">
        <Button asChild>
          <a href={portalUrl} target="_blank" rel="noopener noreferrer">
            Enable Intents
          </a>
        </Button>
        <Button
          variant="outline"
          onClick={() => query.refetch()}
          disabled={query.isFetching}
        >
          {query.isFetching ? "Checking..." : "Check Again"}
        </Button>
      </CardFooter>
    </Card>
  );
}
