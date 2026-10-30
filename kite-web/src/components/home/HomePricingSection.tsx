import PlanCards from "@/components/common/PlanCards";
import PlanIntervalToggle from "@/components/common/PlanIntervalToggle";
import { Button } from "@/components/ui/button";
import { useBillingPlansQuery } from "@/lib/api/queries";
import type { Plans } from "@/lib/apiData";
import { BillingInterval, plansHaveYearly } from "@/lib/billing";
import env from "@/lib/env/client";
import { useResponseData } from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import Link from "next/link";
import { useState } from "react";

// usePublicPlans returns the plans that are shown. They come with the page on
// kite.onl. The export for self-hosting, or a page built without the API,
// loads them in the browser. failed means there's nothing to show, as the API
// can't be reached.
export function usePublicPlans(initial: Plans) {
  const query = useBillingPlansQuery(!initial);
  const plans = (useResponseData(query) ?? initial)?.filter(
    (plan): plan is BillingPlan => !!plan && !plan.hidden
  );
  return { plans, failed: !plans && (query.isError || query.isSuccess) };
}

export default function HomePricingSection({ initial }: { initial: Plans }) {
  const { plans, failed } = usePublicPlans(initial);
  const [interval, setBillingInterval] = useState<BillingInterval>("monthly");
  const yearlyAvailable = !!plans && plansHaveYearly(plans);

  if (failed) return null;

  return (
    <section id="pricing" className="container py-16 sm:py-24 scroll-mt-16">
      <h2 className="text-3xl md:text-4xl font-bold text-center">
        Simple{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Pricing
        </span>
      </h2>
      <p className="text-xl text-muted-foreground text-center mt-4 max-w-2xl mx-auto">
        Every bot starts on the free plan, which keeps it online 24/7 with your
        own name, avatar and status. Upgrade a single bot when it needs higher
        limits, and cancel any time.
      </p>

      {yearlyAvailable && (
        <div className="mt-8">
          <PlanIntervalToggle
            plans={plans}
            value={interval}
            onChange={setBillingInterval}
          />
        </div>
      )}

      <div className="max-w-5xl mx-auto mt-12">
        {plans ? (
          <PlanCards
            plans={plans}
            interval={interval}
            button={(plan) => (
              <Button
                asChild
                variant={plan.popular ? "default" : "outline"}
                className="w-full"
              >
                <Link href="/apps">
                  {plan.price === 0 ? "Start for Free" : `Get ${plan.title}`}
                </Link>
              </Button>
            )}
          />
        ) : (
          // Keeps the height while loading, so the page doesn't jump.
          <div className="h-[460px] rounded-xl border bg-card" />
        )}
      </div>

      <p className="text-sm text-muted-foreground text-center mt-6">
        Most actions, like sending a message, cost 1 credit.{" "}
        <a
          href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
          className="underline hover:text-foreground"
        >
          About credits
        </a>
      </p>

      <div className="text-center mt-4">
        <Link
          href="/pricing#compare"
          className="text-sm font-medium text-primary hover:underline"
        >
          Compare all limits
        </Link>
      </div>
    </section>
  );
}
