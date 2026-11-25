import { BillingPlan, BillingSale } from "./types/wire.gen";

export type BillingInterval = "monthly" | "yearly";

// Yearly is only offered once every paid plan has a yearly price, so the
// toggle never switches some plans and silently leaves others monthly. Plans
// only have one if checkout can sell them yearly.
export function plansHaveYearly(plans: BillingPlan[]) {
  const paid = plans.filter((plan) => plan.price > 0);
  return paid.length > 0 && paid.every((plan) => plan.yearly_price > 0);
}

// The price per month to show big, and for yearly what's actually charged. A
// sale lowers the first year, which is what the big price shows then.
export function planPrice(
  plan: BillingPlan,
  interval: BillingInterval,
  sale?: BillingSale
) {
  if (interval === "yearly" && plan.price > 0) {
    const firstYear = sale ? saleFirstYear(plan, sale) : undefined;
    return {
      perMonth: formatPrice((firstYear ?? plan.yearly_price) / 12),
      billedYearly: formatPrice(plan.yearly_price),
      firstYear: firstYear !== undefined ? formatPrice(firstYear) : undefined,
    };
  }
  return { perMonth: formatPrice(plan.price) };
}

// saleFirstYear is what the first year of the plan costs with the sale,
// rounded to cents like Lemon Squeezy does.
export function saleFirstYear(plan: BillingPlan, sale: BillingSale) {
  return Math.round(plan.yearly_price * (100 - sale.percent)) / 100;
}

// saleLastDay is the last day of the sale, like "November 30".
export function saleLastDay(sale: BillingSale) {
  return new Date(Date.parse(sale.ends_at) - 1).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });
}

function formatPrice(price: number) {
  const rounded = Math.round(price * 100) / 100;
  return Number.isInteger(rounded) ? `${rounded}` : rounded.toFixed(2);
}
