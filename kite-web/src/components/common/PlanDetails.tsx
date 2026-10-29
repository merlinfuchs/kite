import PlanPrice from "@/components/common/PlanPrice";
import { BillingInterval } from "@/lib/billing";
import { planLimits } from "@/lib/planLimits";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { CheckIcon, MinusIcon } from "lucide-react";

// PlanDetails is a table of every limit of every plan, one column per plan.
export default function PlanDetails({
  plans,
  interval,
}: {
  plans: BillingPlan[];
  interval: BillingInterval;
}) {
  return (
    <div className="rounded-xl border bg-card overflow-x-auto">
      <table className="w-full min-w-[640px] text-sm">
        <thead>
          <tr>
            <th className="sticky left-0 bg-card" />
            {plans.map((plan) => (
              <th
                key={plan.id}
                className={cn(
                  "px-4 py-4 text-left font-normal align-bottom",
                  plan.popular && "bg-primary/10"
                )}
              >
                <div className="font-semibold text-base">{plan.title}</div>
                {plan.price === 0 ? (
                  <span className="text-lg font-bold">Free</span>
                ) : (
                  <PlanPrice
                    plan={plan}
                    interval={interval}
                    className="text-lg font-bold"
                    suffix="/ month"
                  />
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {planLimits.map((limit) => (
            <tr key={limit.label} className="border-t">
              <th className="sticky left-0 bg-card px-4 py-2.5 text-left font-normal text-muted-foreground">
                {limit.label}
              </th>
              {plans.map((plan) => {
                const value = limit.value(plan);
                return (
                  <td
                    key={plan.id}
                    className={cn(
                      "px-4 py-2.5 font-medium",
                      plan.popular && "bg-primary/10"
                    )}
                  >
                    {value === true ? (
                      <CheckIcon className="w-5 h-5 text-primary" />
                    ) : value === false ? (
                      <MinusIcon className="w-5 h-5 text-muted-foreground/50" />
                    ) : (
                      value
                    )}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
