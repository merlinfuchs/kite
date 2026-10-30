import { useUpsellStateStore } from "@/lib/state/upsellPopup";
import { CheckIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { useAppId } from "@/lib/hooks/params";
import env from "@/lib/env/client";
import { track } from "@/lib/analytics";
import { useAppPlan, useBillingPlans } from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import { planHighlights } from "@/lib/planLimits";
import { planPrice } from "@/lib/billing";
import { formatNumber } from "@/lib/utils";

export default function UpsellPopup() {
  const shouldUpsell = useUpsellStateStore((s) => s.shouldUpsell);
  const initializeStore = useUpsellStateStore((s) => s.initialize);
  const closeUpsell = useUpsellStateStore((s) => s.closeUpsell);

  const [showUpsell, setShowUpsell] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      setShowUpsell(shouldUpsell());
    }, 5000);

    return () => {
      clearInterval(interval);
    };
  }, [shouldUpsell]);

  useEffect(() => {
    initializeStore();
  }, [initializeStore]);

  const appId = useAppId();
  // Only apps on the free plan are asked, about the cheapest paid plan.
  const current = useAppPlan();
  const plans = useBillingPlans();
  const plan = plans
    ?.filter((p): p is BillingPlan => !!p && !p.hidden && p.price > 0)
    .sort((a, b) => a.price - b.price)[0];
  const visible = showUpsell && !!current?.default && !!plan;

  useEffect(() => {
    if (visible) track("upsell_shown", { from: "popup" });
  }, [visible]);

  if (!visible || !current || !plan) {
    return null;
  }

  const perks = [
    ...(plan.feature_max_commands > current.feature_max_commands
      ? [`${formatNumber(plan.feature_max_commands)} commands`]
      : []),
    ...(plan.feature_usage_credits_per_month >
    current.feature_usage_credits_per_month
      ? [
          `${formatNumber(
            plan.feature_usage_credits_per_month
          )} credits a month`,
        ]
      : []),
    ...planHighlights(plan, current),
  ].slice(0, 3);

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto bottom-[calc(5rem+env(safe-area-inset-bottom))] md:inset-x-auto md:bottom-5 md:right-5 z-50 p-5 space-y-4">
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-1 right-1"
        onClick={closeUpsell}
        aria-label="Close for a week"
      >
        <XIcon className="w-5 h-5" />
      </Button>
      <div>
        <div className="font-semibold text-lg">
          Do more with Kite {plan.title}
        </div>
        <div className="text-sm text-muted-foreground">
          ${planPrice(plan, "monthly").perMonth} a month
        </div>
      </div>
      <ul className="space-y-2 text-sm">
        {perks.map((perk) => (
          <li key={perk} className="flex items-center gap-2">
            <CheckIcon className="size-4 text-primary shrink-0" />
            {perk}
          </li>
        ))}
      </ul>
      <Button asChild className="w-full">
        <Link
          href={`/apps/${appId}/premium`}
          onClick={() => track("upsell_clicked", { from: "popup" })}
        >
          See plans
        </Link>
      </Button>
      <div className="text-xs text-muted-foreground text-center">
        Enjoying Kite?{" "}
        <Link
          href={env.NEXT_PUBLIC_GITHUB_LINK}
          target="_blank"
          className="text-primary hover:underline"
        >
          Star it on GitHub
        </Link>
      </div>
    </Card>
  );
}
