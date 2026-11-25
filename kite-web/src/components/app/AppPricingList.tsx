import PlanCards from "../common/PlanCards";
import Link from "next/link";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import {
  useAppSubscriptions,
  useBillingPlans,
  useBillingSale,
  useSharedBot,
} from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import { useMemo, useState } from "react";
import { useCheckout, useSubscriptionPlanSwitch } from "@/lib/hooks/billing";
import ConfirmDialog from "../common/ConfirmDialog";
import PlanIntervalToggle from "../common/PlanIntervalToggle";
import { BillingInterval, plansHaveYearly, saleLastDay } from "@/lib/billing";

export default function AppPricingList() {
  const subscriptions = useAppSubscriptions();
  // Apps on the Kite bot only use the credits and look of a plan.
  const sharedBot = useSharedBot();

  const activeSubscriptions = useMemo(
    () => subscriptions?.filter((subscription) => subscription!.active),
    [subscriptions]
  );

  const plans = useBillingPlans();
  const visiblePlans = useMemo(
    () =>
      plans?.filter((plan): plan is BillingPlan => !!plan && !plan.hidden) ??
      [],
    [plans]
  );
  const yearlyAvailable = plansHaveYearly(visiblePlans);
  const sale = useBillingSale();

  // Monthly by default, unless the app already pays yearly, so its current
  // plan shows as such, or a sale runs, which is only on yearly plans.
  const [chosenInterval, setChosenInterval] = useState<BillingInterval>();
  const paysYearly = activeSubscriptions?.some(
    (subscription) => subscription!.interval === "yearly"
  );
  const interval: BillingInterval = !yearlyAvailable
    ? "monthly"
    : chosenInterval ?? (paysYearly || sale ? "yearly" : "monthly");

  const pricings = useMemo(() => {
    return visiblePlans.map((plan) => {
      return {
        ...plan,
        current: plan.default
          ? activeSubscriptions?.length === 0
          : activeSubscriptions?.some(
              (subscription) =>
                subscription!.plan_id === plan.id &&
                // Switching between monthly and yearly is a plan switch too.
                (!yearlyAvailable || subscription!.interval === interval)
            ),
      };
    });
  }, [activeSubscriptions, visiblePlans, yearlyAvailable, interval]);

  // Switching plans updates the existing subscription instead of buying a
  // second one. We can only do that for a subscription the current user owns.
  const switchableSubscription = activeSubscriptions?.find(
    (subscription) => subscription!.manageable
  );
  const blockedByOtherOwner =
    !switchableSubscription && !!activeSubscriptions?.length;

  const checkout = useCheckout();
  const switchPlan = useSubscriptionPlanSwitch(switchableSubscription?.id);

  return (
    <div>
      {sale && yearlyAvailable && (
        <div className="mb-8 rounded-xl border border-primary/50 bg-primary/5 p-5">
          <div className="font-medium">
            {sale.name}: {sale.percent}% off the first year of yearly plans
            until {saleLastDay(sale)}
          </div>
          <div className="text-sm text-muted-foreground mt-1">
            The code {sale.code} is applied when you check out a yearly plan.
            {activeSubscriptions?.length && !paysYearly ? (
              <>
                {" "}
                On a monthly plan? Switch to yearly, then email{" "}
                <a
                  href="mailto:contact@kite.onl"
                  className="text-foreground underline"
                >
                  contact@kite.onl
                </a>{" "}
                and we&apos;ll refund {sale.percent}% of the yearly price.
              </>
            ) : null}
          </div>
        </div>
      )}
      {yearlyAvailable && (
        <div className="mb-8">
          <PlanIntervalToggle
            plans={visiblePlans}
            value={interval}
            onChange={setChosenInterval}
            sale={sale}
          />
        </div>
      )}
      <PlanCards
        plans={pricings}
        interval={interval}
        sale={sale}
        kiteBot={sharedBot}
        badge={(pricing) =>
          pricing.current && <Badge className="text-sm">Current</Badge>
        }
        button={(pricing) => (
          <PricingButton
            pricing={pricing}
            interval={yearlyAvailable ? interval : undefined}
            blockedByOtherOwner={blockedByOtherOwner}
            canSwitch={!!switchableSubscription && !pricing.current}
            onCheckout={() => checkout(pricing, interval)}
            onSwitch={() => switchPlan(pricing, interval)}
          />
        )}
      />
      {!sharedBot && (
        <div className="text-center mt-6">
          <Link
            href="/pricing#compare"
            className="text-sm font-medium text-primary hover:underline"
          >
            Compare all limits
          </Link>
        </div>
      )}
    </div>
  );
}

// PricingButton renders the same button either way; switching a plan only wraps
// it in a confirmation, because it charges the customer straight away.
function PricingButton({
  pricing,
  interval,
  canSwitch,
  blockedByOtherOwner,
  onCheckout,
  onSwitch,
}: {
  pricing: {
    title: string;
    price: number;
    popular: boolean;
    current?: boolean;
  };
  // Only named in the confirmation when there's a choice of interval.
  interval?: BillingInterval;
  canSwitch: boolean;
  blockedByOtherOwner: boolean;
  onCheckout: () => void;
  onSwitch: () => void;
}) {
  const button = (
    <Button
      className="w-full"
      disabled={
        pricing.price === 0 ||
        (!canSwitch && (pricing.current || blockedByOtherOwner))
      }
      variant={pricing.popular ? "default" : "outline"}
      onClick={canSwitch ? undefined : onCheckout}
    >
      {canSwitch
        ? "Switch plan"
        : pricing.current
        ? "Current plan"
        : blockedByOtherOwner
        ? "Managed by another user"
        : "Get started"}
    </Button>
  );

  if (!canSwitch) return button;

  return (
    <ConfirmDialog
      title={`Switch to ${pricing.title}${interval ? ` (${interval})` : ""}?`}
      description={
        "Your subscription changes to this plan straight away. " +
        "Our payment provider charges or credits you the prorated difference for the rest of the current billing period."
      }
      onConfirm={onSwitch}
    >
      {button}
    </ConfirmDialog>
  );
}
