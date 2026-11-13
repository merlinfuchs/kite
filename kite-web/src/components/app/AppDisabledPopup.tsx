import { useApp } from "@/lib/hooks/api";
import { canManageSettings } from "@/lib/collaborators";
import { formatMonthlyResetDate } from "@/lib/utils";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { OctagonAlertIcon } from "lucide-react";
import { ReactNode } from "react";
import { Button } from "../ui/button";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";
import GetCreditsButton from "./GetCreditsButton";

export default function AppDisabledPopup() {
  const app = useApp();

  if (!app) return null;
  // Apps paused for credits stay enabled and reply to users. A stopped app is
  // offline, so it only shows why it was stopped.
  const paused = app.enabled && app.no_credits;
  if (!paused && (app.enabled || !app.disabled_reason)) return null;

  let title = "App disabled";
  let description = `Your app is currently disabled: ${app.disabled_reason}`;
  let action: ReactNode = null;
  if (app.token_invalid) {
    title = "Your bot token stopped working";
    if (canManageSettings(app.user_role)) {
      description =
        "The token was probably reset in the Discord Developer Portal. Update it to start your bot again.";
      action = (
        <AppTokenUpdateDialog app={app} from="popup">
          <Button size="sm">Update token</Button>
        </AppTokenUpdateDialog>
      );
    } else {
      description =
        "The token was probably reset in the Discord Developer Portal. Ask an admin or the owner of the app to update it.";
    }
  } else if (paused) {
    title = "Your app is out of credits";
    description = `Your app used all of its credits for this month, so its commands and events are paused. Anyone who uses it is told why. It runs again by itself on ${formatMonthlyResetDate()}, or a few minutes after you upgrade.`;
    action = <GetCreditsButton from="no_credits_popup" />;
  }

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-red-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2">
          <OctagonAlertIcon className="w-5 h-5 text-red-500" />
          {title}
        </CardTitle>
        <CardDescription className="text-sm">{description}</CardDescription>
        {action && <div className="pt-2">{action}</div>}
      </CardHeader>
    </Card>
  );
}
