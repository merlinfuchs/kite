import { track } from "@/lib/analytics";
import { planPrice } from "@/lib/billing";
import {
  useAppFeatures,
  useAppPlan,
  useAppSubscriptions,
  useBillingPlans,
} from "@/lib/hooks/api";
import { useLemonSqueezyCheckout } from "@/lib/hooks/lemonsqueezy";
import { useAppId } from "@/lib/hooks/params";
import { planHighlights } from "@/lib/planLimits";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn, formatNumber } from "@/lib/utils";
import { ArrowRightIcon, CheckIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode, useState } from "react";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";

type LimitKey =
  | "max_commands"
  | "max_event_listeners"
  | "max_scheduled_event_listeners"
  | "max_webhook_event_listeners"
  | "max_messages"
  | "max_variables";

// From 12 of 15 on, the count links to the plans.
const nearLimitRatio = 0.8;

export interface PlanLimit {
  count: number;
  limit: number;
  reached: boolean;
  near: boolean;
  // The cheapest plan with a higher limit, and that limit (0 is unlimited).
  nextPlan?: BillingPlan;
  nextLimit: number;
}

// usePlanLimit returns undefined while loading and when the plan has no limit.
export function usePlanLimit(
  key: LimitKey,
  count: number | undefined
): PlanLimit | undefined {
  const features = useAppFeatures();
  const plans = useBillingPlans();
  if (!features || count === undefined) return undefined;

  const limit = features[key];
  if (!limit) return undefined;

  const planKey = `feature_${key}` as const;
  const nextPlan = plans
    ?.filter(
      (p): p is BillingPlan =>
        !!p && !p.hidden && (p[planKey] === 0 || p[planKey] > limit)
    )
    .sort((a, b) => a.price - b.price)[0];

  return {
    count,
    limit,
    reached: count >= limit,
    near: count >= limit * nearLimitRatio,
    nextPlan,
    nextLimit: nextPlan?.[planKey] ?? 0,
  };
}

function nextLimitText(limit: PlanLimit, noun: string) {
  return limit.nextLimit === 0
    ? `Unlimited ${noun}`
    : `${formatNumber(limit.nextLimit)} ${noun}`;
}

export function PlansLink({
  from,
  className,
  newTab,
  children,
}: {
  from: string;
  className?: string;
  // In the flow editor, so leaving the page doesn't lose unsaved changes.
  newTab?: boolean;
  children: ReactNode;
}) {
  const appId = useAppId();

  return (
    <Link
      href={{ pathname: "/apps/[appId]/premium", query: { appId } }}
      className={className}
      target={newTab ? "_blank" : undefined}
      onClick={() => track("upsell_clicked", { from })}
    >
      {children}
    </Link>
  );
}

function LimitMeter({
  limit,
  className,
}: {
  limit: PlanLimit;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "h-1.5 rounded-full bg-muted overflow-hidden shrink-0",
        className
      )}
    >
      <div
        className={cn(
          "h-full rounded-full",
          limit.reached
            ? "bg-red-500"
            : limit.near
            ? "bg-amber-500"
            : "bg-muted-foreground/60"
        )}
        style={{
          width: `${Math.min(100, (limit.count / limit.limit) * 100)}%`,
        }}
      />
    </div>
  );
}

// PlanLimitCount shows how much of a limit is used, like "12 of 15
// commands", and links to the plans once it's close.
export function PlanLimitCount({
  limit,
  noun,
  from,
}: {
  limit?: PlanLimit;
  noun: string;
  from: string;
}) {
  if (!limit) return null;

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
      <LimitMeter limit={limit} className="w-16" />
      <span className="whitespace-nowrap">
        {limit.count > limit.limit
          ? `${limit.count} ${noun}, your plan allows ${limit.limit}`
          : `${limit.count} of ${limit.limit} ${noun} used`}
      </span>
      {limit.near && limit.nextPlan && (
        <PlansLink
          from={`${from}_list`}
          className="ml-auto inline-flex items-center gap-1 whitespace-nowrap text-primary hover:underline"
        >
          {nextLimitText(limit, noun)} with Kite {limit.nextPlan.title}
          <ArrowRightIcon className="size-3.5" />
        </PlansLink>
      )}
    </div>
  );
}

// PlanLimitGate opens the limit dialog instead of the create or import
// dialog in its children once the limit is reached. It catches the click
// before it reaches their trigger, so those dialogs don't need to know about
// limits.
export function PlanLimitGate({
  limit,
  noun,
  from,
  children,
}: {
  limit?: PlanLimit;
  noun: string;
  from: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  if (!limit?.reached) return <>{children}</>;

  return (
    <>
      <div
        className="contents"
        onClickCapture={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setOpen(true);
          track("upsell_shown", { from: `${from}_limit` });
        }}
      >
        {children}
      </div>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              You&apos;ve used all {limit.limit} {noun}
            </DialogTitle>
            <DialogDescription>
              {limit.nextPlan
                ? `Upgrade to Kite ${limit.nextPlan.title} to keep building, or delete ${noun} you don't need.`
                : `Delete ${noun} you don't need to make room for new ones.`}
            </DialogDescription>
          </DialogHeader>
          {limit.nextPlan && (
            <NextPlanCard limit={limit} plan={limit.nextPlan} noun={noun} />
          )}
          <DialogFooter className="gap-2">
            {limit.nextPlan ? (
              <>
                <Button variant="outline" asChild>
                  <PlansLink from={`${from}_limit`}>
                    Compare all plans
                  </PlansLink>
                </Button>
                <UpgradeButton
                  plan={limit.nextPlan}
                  from={`${from}_limit`}
                  onCheckout={() => setOpen(false)}
                />
              </>
            ) : (
              <Button variant="outline" onClick={() => setOpen(false)}>
                Close
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

// UpgradeButton opens the checkout of the plan right away. Apps that already
// pay go to the plans page instead, where switching plans asks to confirm the
// prorated charge.
export function UpgradeButton({
  plan,
  from,
  onCheckout,
  size,
  newTab,
  children,
}: {
  plan: BillingPlan;
  from: string;
  onCheckout?: () => void;
  size?: "sm";
  newTab?: boolean;
  children?: ReactNode;
}) {
  const subscriptions = useAppSubscriptions();
  const checkout = useLemonSqueezyCheckout();

  const label = children ?? `Upgrade to Kite ${plan.title}`;

  const subscribed = subscriptions?.some((s) => s!.active);
  if (subscriptions && !subscribed) {
    return (
      <Button
        size={size}
        onClick={() => {
          track("upsell_clicked", { from });
          checkout(plan, "monthly");
          onCheckout?.();
        }}
      >
        {label}
      </Button>
    );
  }

  return (
    <Button size={size} asChild>
      <PlansLink from={from} newTab={newTab}>
        {label}
      </PlansLink>
    </Button>
  );
}

// NextPlanCard shows the limit that was reached on the next plan, then a few
// other things the plan adds over the current one.
function NextPlanCard({
  limit,
  plan,
  noun,
}: {
  limit: PlanLimit;
  plan: BillingPlan;
  noun: string;
}) {
  const current = useAppPlan();

  const extras = [
    ...(current &&
    plan.feature_usage_credits_per_month >
      current.feature_usage_credits_per_month
      ? [
          `${formatNumber(
            plan.feature_usage_credits_per_month
          )} credits a month`,
        ]
      : []),
    // The highlights can name the limit that was reached again.
    ...planHighlights(plan, current).filter((h) => !h.includes(noun)),
  ].slice(0, 3);

  return (
    <div className="rounded-xl border border-primary/30 bg-primary/5 p-4 space-y-3">
      <div className="flex items-baseline justify-between gap-3">
        <div className="font-semibold">{plan.title}</div>
        <div>
          <span className="text-lg font-bold">
            ${planPrice(plan, "monthly").perMonth}
          </span>
          <span className="text-sm text-muted-foreground"> / month</span>
        </div>
      </div>
      <ul className="space-y-2 text-sm">
        <li className="flex items-center gap-2">
          <CheckIcon className="size-4 text-primary shrink-0" />
          <span className="font-medium">{nextLimitText(limit, noun)}</span>
          <span className="text-muted-foreground">
            instead of {limit.limit}
          </span>
        </li>
        {extras.map((extra) => (
          <li key={extra} className="flex items-center gap-2">
            <CheckIcon className="size-4 text-primary shrink-0" />
            {extra}
          </li>
        ))}
      </ul>
    </div>
  );
}
