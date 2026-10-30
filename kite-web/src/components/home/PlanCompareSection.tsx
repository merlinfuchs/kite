import PlanDetails from "@/components/common/PlanDetails";
import PlanIntervalToggle from "@/components/common/PlanIntervalToggle";
import type { Plans } from "@/lib/apiData";
import { BillingInterval, plansHaveYearly } from "@/lib/billing";
import { useState } from "react";
import { usePublicPlans } from "./HomePricingSection";

// Every limit of every plan, with its own billing interval so it can be
// anywhere on the page.
export default function PlanCompareSection({ initial }: { initial: Plans }) {
  const { plans } = usePublicPlans(initial);
  const [interval, setBillingInterval] = useState<BillingInterval>("monthly");

  if (!plans) return null;

  return (
    <section id="compare" className="container py-12 md:py-16 scroll-mt-20">
      <div className="max-w-5xl mx-auto">
        <div className="flex flex-wrap items-end justify-between gap-4 mb-6">
          <div>
            <h2 className="text-2xl md:text-3xl font-bold">
              Compare the Plans
            </h2>
            <p className="text-muted-foreground mt-2">
              Every limit of every plan. All plans keep your bot online 24/7
              with your own name, avatar and status.
            </p>
          </div>
          {plansHaveYearly(plans) && (
            <PlanIntervalToggle
              plans={plans}
              value={interval}
              onChange={setBillingInterval}
            />
          )}
        </div>
        <PlanDetails plans={plans} interval={interval} />
      </div>
    </section>
  );
}
