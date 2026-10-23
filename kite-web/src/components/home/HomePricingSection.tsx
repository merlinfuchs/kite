import PlanFeatureList from "@/components/common/PlanFeatureList";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useBillingPlansQuery } from "@/lib/api/queries";
import env from "@/lib/env/client";
import { useResponseData } from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import Link from "next/link";

export default function HomePricingSection() {
  const query = useBillingPlansQuery();
  const plans = useResponseData(query)?.filter(
    (plan): plan is BillingPlan => !!plan && !plan.hidden
  );

  // Nothing to show if the API can't be reached.
  if (query.isError || (query.isSuccess && !plans)) return null;

  return (
    <section id="pricing" className="container py-16 sm:py-24 scroll-mt-16">
      <h2 className="text-3xl md:text-4xl font-bold text-center">
        Simple{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Pricing
        </span>
      </h2>
      <p className="text-xl text-muted-foreground text-center mt-4 max-w-2xl mx-auto">
        Every bot starts on the free plan. Upgrade a single bot when it needs
        higher limits, and cancel any time.
      </p>

      <div className="grid md:grid-cols-3 gap-6 mt-12 max-w-5xl mx-auto">
        {plans
          ? plans.map((plan) => (
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
                  {plan.price === 0 ? (
                    <span className="text-4xl font-bold">Free</span>
                  ) : (
                    <>
                      <span className="text-4xl font-bold">${plan.price}</span>
                      <span className="text-muted-foreground">
                        {" "}
                        / month per bot
                      </span>
                    </>
                  )}
                </div>
                <p className="text-muted-foreground mt-2">{plan.description}</p>
                <Button
                  asChild
                  variant={plan.popular ? "default" : "outline"}
                  className="mt-6"
                >
                  <Link href="/apps">
                    {plan.price === 0 ? "Get Started for Free" : "Get Started"}
                  </Link>
                </Button>
                <hr className="my-6" />
                <PlanFeatureList plan={plan} />
              </div>
            ))
          : // Keeps the height while loading, so the page doesn't jump.
            Array.from({ length: 3 }, (_, i) => (
              <div key={i} className="h-[620px] rounded-xl border bg-card" />
            ))}
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
    </section>
  );
}
