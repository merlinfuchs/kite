import { useApp } from "@/lib/hooks/api";
import { formatMonthlyResetDate } from "@/lib/utils";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { OctagonAlertIcon } from "lucide-react";
import { ReactNode } from "react";
import { Button } from "../ui/button";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";
import GetCreditsButton from "./GetCreditsButton";

export default function AppDisabledPopup() {
  const app = useApp();

  if (app?.enabled || !app?.disabled_reason) return null;

  let title = "App Disabled";
  let description = `Your app is currently disabled: ${app.disabled_reason}`;
  let action: ReactNode = null;
  if (app.token_invalid) {
    title = "Your Bot Token Stopped Working";
    description =
      "The token was probably reset in the Discord Developer Portal. Update it to start your bot again.";
    action = (
      <AppTokenUpdateDialog app={app} from="popup">
        <Button size="sm">Update token</Button>
      </AppTokenUpdateDialog>
    );
  } else if (app.no_credits) {
    title = "Your App Is Out of Credits";
    description = `Your bot used all of its credits for this month and is offline. It starts again by itself on ${formatMonthlyResetDate()}, or a few minutes after you upgrade.`;
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
