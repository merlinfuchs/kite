import PlanPrice from "@/components/common/PlanPrice";
import { Badge } from "@/components/ui/badge";
import { BillingInterval } from "@/lib/billing";
import { planLimits } from "@/lib/planLimits";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { CheckIcon, MinusIcon } from "lucide-react";

// PlanDetails lists every limit of every plan, one plan per row. The limits are
// in the same order in each row, so they line up for comparing.
export default function PlanDetails({
  plans,
  interval,
}: {
  plans: BillingPlan[];
  interval: BillingInterval;
}) {
  return (
    <div className="rounded-xl border bg-card divide-y overflow-hidden">
      {plans.map((plan) => (
        <div
          key={plan.id}
          className={cn(
            "p-5 grid lg:grid-cols-[200px_1fr] gap-x-8 gap-y-4",
            plan.popular && "bg-primary/5"
          )}
        >
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-semibold">{plan.title}</h3>
              {plan.popular && (
                <Badge variant="secondary" className="text-primary">
                  Best Value
                </Badge>
              )}
            </div>
            <div className="mt-1">
              {plan.price === 0 ? (
                <span className="text-2xl font-bold">Free</span>
              ) : (
                <PlanPrice
                  plan={plan}
                  interval={interval}
                  className="text-2xl font-bold"
                  suffix="/ month"
                />
              )}
            </div>
          </div>
          <dl className="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-x-6 gap-y-3 text-sm">
            {planLimits.map((limit) => {
              const value = limit.value(plan);
              return (
                <div key={limit.label}>
                  <dt className="text-xs text-muted-foreground">
                    {limit.label}
                  </dt>
                  <dd className="font-medium">
                    {value === true ? (
                      <CheckIcon className="w-5 h-5 text-primary" />
                    ) : value === false ? (
                      <MinusIcon className="w-5 h-5 text-muted-foreground/50" />
                    ) : (
                      value
                    )}
                  </dd>
                </div>
              );
            })}
          </dl>
        </div>
      ))}
    </div>
  );
}
