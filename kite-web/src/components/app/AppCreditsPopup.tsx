import { useApp, useAppFeature, useUsageCredits } from "@/lib/hooks/api";
import { formatMonthlyResetDate, formatNumber } from "@/lib/utils";
import { TriangleAlertIcon, XIcon } from "lucide-react";
import { useState } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import GetCreditsButton from "./GetCreditsButton";

// Shown from this share of the credits on, so there is time to act before the
// app is stopped.
const warnAt = 0.8;

export default function AppCreditsPopup() {
  const app = useApp();
  const credits = useUsageCredits();
  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month);

  // AppShell stays mounted when switching apps, so this is per app.
  const [closedAppId, setClosedAppId] = useState<string>();

  if (!app?.enabled || closedAppId === app.id) return null;
  if (!credits || !creditsPerMonth) return null;
  if (credits.credits_used < creditsPerMonth * warnAt) return null;

  const left = creditsPerMonth - credits.credits_used;

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-yellow-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2 pr-6">
          <TriangleAlertIcon className="w-5 h-5 text-yellow-500" />
          {left > 0
            ? "Your App Is Running Low on Credits"
            : "Your App Is Out of Credits"}
        </CardTitle>
        <CardDescription className="text-sm">
          {left > 0
            ? `${formatNumber(left)} of ${formatNumber(
                creditsPerMonth
              )} credits are left this month. When they run out, your bot goes offline until ${formatMonthlyResetDate()}.`
            : `Your bot used all of its credits for this month and goes offline in a few minutes, until ${formatMonthlyResetDate()}.`}
        </CardDescription>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-1 right-1"
          onClick={() => setClosedAppId(app.id)}
        >
          <XIcon className="w-5 h-5" />
        </Button>
        <div className="pt-2">
          <GetCreditsButton from="low_credits_popup" />
        </div>
      </CardHeader>
    </Card>
  );
}
