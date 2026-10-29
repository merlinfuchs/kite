import PlanFeatureList from "@/components/common/PlanFeatureList";
import PlanIntervalToggle from "@/components/common/PlanIntervalToggle";
import PlanPrice from "@/components/common/PlanPrice";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useBillingPlansQuery } from "@/lib/api/queries";
import type { Plans } from "@/lib/apiData";
import { BillingInterval, plansHaveYearly } from "@/lib/billing";
import env from "@/lib/env/client";
import { useResponseData } from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { useState } from "react";

// The plans come with the page on kite.onl. The export for self-hosting, or a
// page built without the API, loads them in the browser.
export default function HomePricingSection({ initial }: { initial: Plans }) {
  const query = useBillingPlansQuery(!initial);
  const plans = (useResponseData(query) ?? initial)?.filter(
    (plan): plan is BillingPlan => !!plan && !plan.hidden
  );
  const [interval, setBillingInterval] = useState<BillingInterval>("monthly");
  const yearlyAvailable = !!plans && plansHaveYearly(plans);
  // The free plan is a row of its own above the paid ones, so they stay a
  // choice of three.
  const freePlan = plans?.find((plan) => plan.price === 0);
  const paidPlans = plans?.filter((plan) => plan.price !== 0) ?? [];

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

      {plans ? (
        <div className="max-w-5xl mx-auto mt-12 space-y-6">
          {freePlan && (
            <div className="rounded-xl border bg-card p-6 flex flex-col lg:flex-row gap-6 lg:gap-10">
              <div className="lg:w-64 flex-none flex flex-col">
                <h3 className="text-xl font-semibold">{freePlan.title}</h3>
                <span className="text-4xl font-bold mt-4">Free</span>
                <p className="text-muted-foreground mt-2">
                  {freePlan.description}
                </p>
                <Button asChild variant="outline" className="mt-6">
                  <Link href="/apps">Get Started for Free</Link>
                </Button>
              </div>
              <PlanFeatureList
                plan={freePlan}
                className="grid sm:grid-cols-2 gap-x-8 gap-y-3 space-y-0 content-center flex-1"
              />
            </div>
          )}

          <div
            className={cn(
              "grid gap-6",
              paidPlans.length > 2 ? "md:grid-cols-3" : "md:grid-cols-2"
            )}
          >
            {paidPlans.map((plan) => (
              <div
                key={plan.id}
                className={cn(
                  "rounded-xl border bg-card p-6 flex flex-col",
                  plan.popular && "border-2 border-primary shadow-lg"
                )}
              >
                <div className="flex items-center justify-between">
                  <h3 className="text-xl font-semibold">{plan.title}</h3>
                  {plan.popular && (
                    <Badge variant="secondary" className="text-primary">
                      Best Value
                    </Badge>
                  )}
                </div>
                <div className="mt-4">
                  <PlanPrice
                    plan={plan}
                    interval={interval}
                    className="text-4xl font-bold"
                    suffix="/ month per bot"
                  />
                </div>
                <p className="text-muted-foreground mt-2">{plan.description}</p>
                <Button
                  asChild
                  variant={plan.popular ? "default" : "outline"}
                  className="mt-6"
                >
                  <Link href="/apps">Get Started</Link>
                </Button>
                <hr className="my-6" />
                <PlanFeatureList plan={plan} />
              </div>
            ))}
          </div>
        </div>
      ) : (
        // Keeps the height while loading, so the page doesn't jump.
        <div className="max-w-5xl mx-auto mt-12 space-y-6">
          <div className="h-[260px] rounded-xl border bg-card" />
          <div className="grid md:grid-cols-3 gap-6">
            {Array.from({ length: 3 }, (_, i) => (
              <div key={i} className="h-[620px] rounded-xl border bg-card" />
            ))}
          </div>
        </div>
      )}

      <p className="text-sm text-muted-foreground text-center mt-6">
        Most actions, like sending a message, cost 1 credit.{" "}
        <a
          href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
          className="underline hover:text-foreground"
        >
          About credits
        </a>
      </p>
    </section>
  );
}
