import PlanFeatureList from "../common/PlanFeatureList";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Badge } from "../ui/badge";
import { useAppSubscriptions, useBillingPlans } from "@/lib/hooks/api";
import { BillingPlan } from "@/lib/types/wire.gen";
import { useMemo, useState } from "react";
import {
  useLemonSqueezyCheckout,
  useSubscriptionPlanSwitch,
} from "@/lib/hooks/lemonsqueezy";
import ConfirmDialog from "../common/ConfirmDialog";
import PlanIntervalToggle from "../common/PlanIntervalToggle";
import PlanPrice from "../common/PlanPrice";
import { BillingInterval, plansHaveYearly, planVariantId } from "@/lib/billing";

export default function AppPricingList() {
  const subscriptions = useAppSubscriptions();

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

  // Monthly by default, unless the app already pays yearly, so its current
  // plan shows as such.
  const [chosenInterval, setChosenInterval] = useState<BillingInterval>();
  const paysYearly = activeSubscriptions?.some((subscription) =>
    visiblePlans.some(
      (plan) =>
        planVariantId(plan, "yearly") === subscription!.lemonsqueezy_variant_id
    )
  );
  const interval: BillingInterval = !yearlyAvailable
    ? "monthly"
    : chosenInterval ?? (paysYearly ? "yearly" : "monthly");

  const pricings = useMemo(() => {
    return visiblePlans.map((plan) => {
      return {
        ...plan,
        current: plan.default
          ? activeSubscriptions?.length === 0
          : activeSubscriptions?.some(
              (subscription) =>
                subscription!.lemonsqueezy_product_id ===
                  plan.lemonsqueezy_product_id &&
                // Monthly and yearly are variants of the same product, so
                // switching between them is a plan switch too.
                (!yearlyAvailable ||
                  subscription!.lemonsqueezy_variant_id ===
                    planVariantId(plan, interval))
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

  const checkout = useLemonSqueezyCheckout();
  const switchPlan = useSubscriptionPlanSwitch(switchableSubscription?.id);

  return (
    <div>
      {yearlyAvailable && (
        <div className="mb-8">
          <PlanIntervalToggle
            plans={visiblePlans}
            value={interval}
            onChange={setChosenInterval}
          />
        </div>
      )}
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-8 xl:mx-16">
        {pricings.map((pricing) => (
          <Card
            key={pricing.title}
            className={
              pricing.popular
                ? "drop-shadow-xl shadow-black/10 dark:shadow-white/10"
                : "xl:my-8 "
            }
          >
            <CardHeader>
              <CardTitle className="flex item-center justify-between">
                {pricing.title}
                <div className="flex gap-2">
                  {pricing.current ? (
                    <Badge className="text-sm">Current</Badge>
                  ) : null}
                  {pricing.popular ? (
                    <Badge variant="secondary" className="text-sm text-primary">
                      Best Value
                    </Badge>
                  ) : null}
                </div>
              </CardTitle>
              <PlanPrice
                plan={pricing}
                interval={interval}
                className="text-3xl font-bold"
                suffix="/month"
              />

              <CardDescription>{pricing.description}</CardDescription>
            </CardHeader>

            <CardContent>
              <PricingButton
                pricing={pricing}
                interval={yearlyAvailable ? interval : undefined}
                blockedByOtherOwner={blockedByOtherOwner}
                canSwitch={!!switchableSubscription && !pricing.current}
                onCheckout={() => checkout(pricing, interval)}
                onSwitch={() => switchPlan(pricing, interval)}
              />
            </CardContent>

            <hr className="w-4/5 m-auto mb-4" />

            <CardFooter className="flex">
              <PlanFeatureList plan={pricing} />
            </CardFooter>
          </Card>
        ))}
      </div>
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
        ? "Switch Plan"
        : pricing.current
        ? "Current Plan"
        : blockedByOtherOwner
        ? "Managed by another user"
        : "Get Started"}
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
