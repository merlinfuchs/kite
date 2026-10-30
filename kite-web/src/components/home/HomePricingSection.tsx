import PlanCards from "@/components/common/PlanCards";
import PlanDetails from "@/components/common/PlanDetails";
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

// The plans come with the page on kite.onl. The export for self-hosting, or a
// page built without the API, loads them in the browser.
// The pricing page passes details to show every limit below the cards, the
// homepage links there instead.
export default function HomePricingSection({
  initial,
  details,
}: {
  initial: Plans;
  details?: boolean;
}) {
  const query = useBillingPlansQuery(!initial);
  const plans = (useResponseData(query) ?? initial)?.filter(
    (plan): plan is BillingPlan => !!plan && !plan.hidden
  );
  const [interval, setBillingInterval] = useState<BillingInterval>("monthly");
  const yearlyAvailable = !!plans && plansHaveYearly(plans);

  // Nothing to show if the API can't be reached.
  if (!plans && (query.isError || query.isSuccess)) return null;

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

      {details ? (
        plans && (
          <div id="compare" className="max-w-5xl mx-auto mt-16 scroll-mt-20">
            <h2 className="text-2xl md:text-3xl font-bold">
              Compare the Plans
            </h2>
            <p className="text-muted-foreground mt-2 mb-6">
              Every limit of every plan. All plans keep your bot online 24/7
              with your own name, avatar and status.
            </p>
            <PlanDetails plans={plans} interval={interval} />
          </div>
        )
      ) : (
        <div className="text-center mt-4">
          <Link
            href="/pricing#compare"
            className="text-sm font-medium text-primary hover:underline"
          >
            Compare all limits
          </Link>
        </div>
      )}
    </section>
  );
}
