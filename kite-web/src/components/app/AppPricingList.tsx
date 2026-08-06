import { CheckIcon, InfinityIcon } from "lucide-react";
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
import { ReactNode, useMemo } from "react";
import {
  useLemonSqueezyCheckout,
  useSubscriptionPlanSwitch,
} from "@/lib/hooks/lemonsqueezy";
import { formatNumber } from "@/lib/utils";
import { isSubscriptionActive } from "@/lib/subscriptions";
import ConfirmDialog from "../common/ConfirmDialog";

export default function AppPricingList() {
  const subscriptions = useAppSubscriptions();

  const activeSubscriptions = subscriptions?.filter((subscription) =>
    isSubscriptionActive(subscription!)
  );

  const plans = useBillingPlans();

  const pricings = useMemo(() => {
    return (
      plans
        ?.filter((plan) => !plan!.hidden)
        .map((plan) => {
          return {
            ...plan!,
            current: activeSubscriptions?.some(
              (subscription) =>
                subscription!.lemonsqueezy_product_id ===
                plan!.lemonsqueezy_product_id
            ),
          };
        }) ?? []
    );
  }, [activeSubscriptions, plans]);

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
    <div className="grid lg:grid-cols-2 xl:grid-cols-3 gap-8 xl:mx-16">
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
              {pricing.popular ? (
                <Badge variant="secondary" className="text-sm text-primary">
                  Best Value
                </Badge>
              ) : null}
            </CardTitle>
            <div>
              <span className="text-3xl font-bold">${pricing.price}</span>
              <span className="text-muted-foreground"> /month</span>
            </div>

            <CardDescription>{pricing.description}</CardDescription>
          </CardHeader>

          <CardContent>
            {switchableSubscription && !pricing.current ? (
              <ConfirmDialog
                title={`Switch to ${pricing.title}?`}
                description={
                  "Your subscription changes to this plan straight away. " +
                  "Our payment provider charges or credits you the prorated difference for the rest of the current billing period."
                }
                onConfirm={() => switchPlan(pricing.lemonsqueezy_variant_id)}
              >
                <Button
                  className="w-full"
                  disabled={pricing.price === 0}
                  variant={pricing.popular ? "default" : "outline"}
                >
                  Switch Plan
                </Button>
              </ConfirmDialog>
            ) : (
              <Button
                className="w-full"
                disabled={
                  pricing.current || pricing.price === 0 || blockedByOtherOwner
                }
                variant={pricing.popular ? "default" : "outline"}
                onClick={() => checkout(pricing.lemonsqueezy_variant_id)}
              >
                {pricing.current
                  ? "Current Plan"
                  : blockedByOtherOwner
                  ? "Managed by another user"
                  : "Get Started"}
              </Button>
            )}
          </CardContent>

          <hr className="w-4/5 m-auto mb-4" />

          <CardFooter className="flex">
            <div className="space-y-4">
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">
                  {pricing.feature_max_collaborators} Collaborator
                  {pricing.feature_max_collaborators === 1 ? "" : "s"}
                </h3>
              </span>
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">
                  {formatNumber(pricing.feature_usage_credits_per_month)}{" "}
                  Credits / month
                </h3>
              </span>
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">{pricing.feature_max_guilds} Servers</h3>
              </span>
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">
                  {pricing.feature_max_commands} Commands & Variables
                </h3>
              </span>
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">
                  {pricing.feature_max_event_listeners} Event Listeners
                </h3>
              </span>
              <span className="flex">
                <CheckIcon className="text-green-500" />{" "}
                <h3 className="ml-2">
                  {pricing.feature_priority_support
                    ? "Priority support"
                    : "Community support"}
                </h3>
              </span>
            </div>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
