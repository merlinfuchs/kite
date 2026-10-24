import { useApp, useAppFeature, useUsageCredits } from "@/lib/hooks/api";
import { track } from "@/lib/analytics";
import { formatCreditsResetDate, formatNumber } from "@/lib/utils";
import { TriangleAlertIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";

// Shown from this share of the credits on, so there is time to act before the
// app is stopped.
const warnAt = 0.8;

export default function AppCreditsPopup() {
  const app = useApp();
  const credits = useUsageCredits();
  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month);

  const [closed, setClosed] = useState(false);

  if (!app?.enabled || closed || !credits || !creditsPerMonth) return null;
  if (credits.credits_used < creditsPerMonth * warnAt) return null;

  const left = Math.max(creditsPerMonth - credits.credits_used, 0);

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-yellow-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2 pr-6">
          <TriangleAlertIcon className="w-5 h-5 text-yellow-500" />
          Your App Is Running Low on Credits
        </CardTitle>
        <CardDescription className="text-sm">
          {formatNumber(left)} of {formatNumber(creditsPerMonth)} credits are
          left this month. When they run out, your bot goes offline until{" "}
          {formatCreditsResetDate()}.
        </CardDescription>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-1 right-1"
          onClick={() => setClosed(true)}
        >
          <XIcon className="w-5 h-5" />
        </Button>
        <div className="pt-2">
          <Button size="sm" asChild>
            <Link
              href={`/apps/${app.id}/premium`}
              onClick={() =>
                track("upsell_clicked", { from: "low_credits_popup" })
              }
            >
              Get more credits
            </Link>
          </Button>
        </div>
      </CardHeader>
    </Card>
  );
}
