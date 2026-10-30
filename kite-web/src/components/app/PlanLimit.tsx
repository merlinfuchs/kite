import { track } from "@/lib/analytics";
import { useAppFeatures, useBillingPlans } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { BillingPlan } from "@/lib/types/wire.gen";
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
    ? `unlimited ${noun}`
    : `${limit.nextLimit} ${noun}`;
}

function PlansLink({
  from,
  className,
  children,
}: {
  from: string;
  className?: string;
  children: ReactNode;
}) {
  const appId = useAppId();

  return (
    <Link
      href={{ pathname: "/apps/[appId]/premium", query: { appId } }}
      className={className}
      onClick={() => track("upsell_clicked", { from })}
    >
      {children}
    </Link>
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
    <div className="text-sm text-muted-foreground">
      {limit.count > limit.limit
        ? `${limit.count} ${noun}, your plan allows ${limit.limit}`
        : `${limit.count} of ${limit.limit} ${noun}`}
      {limit.near && limit.nextPlan && (
        <>
          {" · "}
          <PlansLink
            from={`${from}_list`}
            className="text-primary hover:underline"
          >
            Get {nextLimitText(limit, noun)} with {limit.nextPlan.title}
          </PlansLink>
        </>
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
              Your plan includes {limit.limit} {noun}
            </DialogTitle>
            <DialogDescription>
              {limit.nextPlan
                ? `${limit.nextPlan.title} gives you ${nextLimitText(
                    limit,
                    noun
                  )} for $${
                    limit.nextPlan.price
                  } a month. Or delete ${noun} you don't need to make room.`
                : `Delete ${noun} you don't need to make room for new ones.`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setOpen(false)}>
              Close
            </Button>
            {limit.nextPlan && (
              <Button asChild>
                <PlansLink from={`${from}_limit`}>See plans</PlansLink>
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
