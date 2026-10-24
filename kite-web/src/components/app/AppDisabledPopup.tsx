import { useApp } from "@/lib/hooks/api";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { OctagonAlertIcon } from "lucide-react";
import { Button } from "../ui/button";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";
import Link from "next/link";
import { formatCreditsResetDate } from "@/lib/utils";
import { track } from "@/lib/analytics";

export default function AppDisabledPopup() {
  const app = useApp();

  if (app?.enabled || !app?.disabled_reason) return null;

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-red-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2">
          <OctagonAlertIcon className="w-5 h-5 text-red-500" />
          {app.token_invalid
            ? "Your Bot Token Stopped Working"
            : app.no_credits
            ? "Your App Is Out of Credits"
            : "App Disabled"}
        </CardTitle>
        <CardDescription className="text-sm">
          {app.token_invalid
            ? "The token was probably reset in the Discord Developer Portal. Update it to start your bot again."
            : app.no_credits
            ? `Your bot used all of its credits for this month and is offline. It starts again by itself on ${formatCreditsResetDate()}, or right away when you upgrade.`
            : `Your app is currently disabled: ${app.disabled_reason}`}
        </CardDescription>
        {app.token_invalid && (
          <div className="pt-2">
            <AppTokenUpdateDialog app={app} from="popup">
              <Button size="sm">Update token</Button>
            </AppTokenUpdateDialog>
          </div>
        )}
        {app.no_credits && (
          <div className="pt-2">
            <Button size="sm" asChild>
              <Link
                href={`/apps/${app.id}/premium`}
                onClick={() =>
                  track("upsell_clicked", { from: "no_credits_popup" })
                }
              >
                Get more credits
              </Link>
            </Button>
          </div>
        )}
      </CardHeader>
    </Card>
  );
}
