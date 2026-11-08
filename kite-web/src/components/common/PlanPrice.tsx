import { BillingInterval, planPrice } from "@/lib/billing";
import { BillingPlan } from "@/lib/types/wire.gen";

export default function PlanPrice({
  plan,
  interval,
  className,
  suffix,
}: {
  plan: BillingPlan;
  interval: BillingInterval;
  className: string;
  suffix: string;
}) {
  const { perMonth, billedYearly } = planPrice(plan, interval);

  return (
    <div>
      <span className={className}>${perMonth}</span>
      <span className="text-muted-foreground"> {suffix}</span>
      {billedYearly && (
        <div className="text-sm text-muted-foreground">
          Billed ${billedYearly} yearly
        </div>
      )}
    </div>
  );
}
