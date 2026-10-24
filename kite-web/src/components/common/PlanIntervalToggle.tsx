import { BillingPlan } from "@/lib/types/wire.gen";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";

export type BillingInterval = "monthly" | "yearly";

// Yearly is only offered once every paid plan has a yearly variant, so the
// toggle never switches some plans and silently leaves others monthly.
export function plansHaveYearly(plans: BillingPlan[]) {
  const paid = plans.filter((plan) => plan.price > 0);
  return (
    paid.length > 0 && paid.every((plan) => plan.lemonsqueezy_yearly_variant_id)
  );
}

export function planVariantId(plan: BillingPlan, interval: BillingInterval) {
  return interval === "yearly"
    ? plan.lemonsqueezy_yearly_variant_id
    : plan.lemonsqueezy_variant_id;
}

// The price per month to show big, and for yearly what's actually charged.
export function planPrice(plan: BillingPlan, interval: BillingInterval) {
  if (interval === "yearly" && plan.price > 0) {
    return {
      perMonth: formatPrice(plan.yearly_price / 12),
      billedYearly: formatPrice(plan.yearly_price),
    };
  }
  return { perMonth: formatPrice(plan.price) };
}

function formatPrice(price: number) {
  const rounded = Math.round(price * 100) / 100;
  return Number.isInteger(rounded) ? `${rounded}` : rounded.toFixed(2);
}

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
