import { BillingPlan } from "./types/wire.gen";

export type BillingInterval = "monthly" | "yearly";

// Yearly is only offered once every paid plan has a yearly variant and price,
// so the toggle never switches some plans and silently leaves others monthly.
export function plansHaveYearly(plans: BillingPlan[]) {
  const paid = plans.filter((plan) => plan.price > 0);
  return (
    paid.length > 0 &&
    paid.every(
      (plan) => plan.lemonsqueezy_yearly_variant_id && plan.yearly_price > 0
    )
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
