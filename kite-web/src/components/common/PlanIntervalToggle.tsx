import { BillingInterval, planPrice } from "@/lib/billing";
import { BillingPlan } from "@/lib/types/wire.gen";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";

export default function PlanIntervalToggle({
  plans,
  value,
  onChange,
}: {
  plans: BillingPlan[];
  value: BillingInterval;
  onChange: (value: BillingInterval) => void;
}) {
  // The smallest saving of any plan, so the badge holds for every plan.
  const saving = Math.min(
    ...plans
      .filter((plan) => plan.price > 0)
      .map((plan) =>
        Math.floor((1 - plan.yearly_price / (plan.price * 12)) * 100)
      )
  );

  return (
    <Tabs
      value={value}
      onValueChange={(v) => onChange(v as BillingInterval)}
      className="flex justify-center"
    >
      <TabsList>
        <TabsTrigger value="monthly" className="px-4">
          Monthly
        </TabsTrigger>
        <TabsTrigger value="yearly" className="px-4 gap-2">
          Yearly
          {saving > 0 && (
            <span className="text-xs font-semibold text-primary">
              Save {saving}%
            </span>
          )}
        </TabsTrigger>
      </TabsList>
    </Tabs>
  );
}

export function PlanPrice({
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
