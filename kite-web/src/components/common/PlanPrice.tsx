import { BillingInterval, planPrice } from "@/lib/billing";
import { BillingPlan, BillingSale } from "@/lib/types/wire.gen";

export default function PlanPrice({
  plan,
  interval,
  sale,
  className,
  suffix,
}: {
  plan: BillingPlan;
  interval: BillingInterval;
  sale?: BillingSale;
  className: string;
  suffix: string;
}) {
  const { perMonth, billedYearly, firstYear } = planPrice(plan, interval, sale);

  return (
    <div>
      <span className={className}>${perMonth}</span>
      <span className="text-muted-foreground"> {suffix}</span>
      {firstYear ? (
        <div className="text-sm text-muted-foreground">
          <s>${billedYearly}</s>{" "}
          <span className="text-foreground">${firstYear}</span> for the first
          year, then ${billedYearly} yearly
        </div>
      ) : (
        billedYearly && (
          <div className="text-sm text-muted-foreground">
            Billed ${billedYearly} yearly
          </div>
        )
      )}
    </div>
  );
}
