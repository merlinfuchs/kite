import PlanPrice from "@/components/common/PlanPrice";
import { Badge } from "@/components/ui/badge";
import { BillingInterval } from "@/lib/billing";
import { planHighlights, shortNumber } from "@/lib/planLimits";
import { BillingPlan } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { CheckIcon } from "lucide-react";
import { ReactNode } from "react";

const columns: Record<number, string> = {
  1: "lg:grid-cols-1",
  2: "lg:grid-cols-2",
  3: "lg:grid-cols-3",
};

// PlanCards are an overview of the plans: the three limits people compare
// first, and what each plan adds over the one before it. PlanDetails has the
// rest. Only the first three plans get a card, the plans after them are a row
// below for the few bots that need more.
export default function PlanCards<P extends BillingPlan>({
  plans,
  interval,
  button,
  badge,
}: {
  plans: P[];
  interval: BillingInterval;
  button: (plan: P) => ReactNode;
  // Shown next to the title, like the current plan of an app.
  badge?: (plan: P) => ReactNode;
}) {
  const cards = plans.slice(0, 3);
  const more = plans.slice(3);
  const previous = (plan: P) => plans[plans.indexOf(plan) - 1];

  return (
    <div className="space-y-4">
      <div
        className={cn(
          "grid md:grid-cols-2 gap-4",
          columns[cards.length] ?? "lg:grid-cols-3"
        )}
      >
        {cards.map((plan) => (
          <div
            key={plan.id}
            className={cn(
              "rounded-xl border bg-card p-5 flex flex-col",
              plan.popular && "border-2 border-primary shadow-lg"
            )}
          >
            <div className="flex items-center justify-between gap-2 min-h-[26px]">
              <h3 className="font-semibold">{plan.title}</h3>
              <div className="flex gap-2">
                {badge?.(plan)}
                {plan.popular && (
                  <Badge variant="secondary" className="text-primary">
                    Best Value
                  </Badge>
                )}
              </div>
            </div>
            <div className="mt-3 min-h-[60px]">
              {plan.price === 0 ? (
                <span className="text-3xl font-bold">Free</span>
              ) : (
                <PlanPrice
                  plan={plan}
                  interval={interval}
                  className="text-3xl font-bold"
                  suffix="/ month"
                />
              )}
            </div>
            <div className="grid grid-cols-3 gap-1.5 mt-3">
              {[
                [shortNumber(plan.feature_usage_credits_per_month), "credits"],
                [shortNumber(plan.feature_max_guilds), "servers"],
                [shortNumber(plan.feature_max_commands), "commands"],
              ].map(([value, label]) => (
                <div
                  key={label}
                  className="rounded-lg bg-muted/50 px-1 py-2 text-center"
                >
                  <div className="text-lg font-bold">{value}</div>
                  <div className="text-xs text-muted-foreground">{label}</div>
                </div>
              ))}
            </div>
            <ul className="mt-5 space-y-2 text-sm text-muted-foreground flex-1">
              {previous(plan) && (
                <li className="text-foreground">
                  Everything in {previous(plan).title}, and:
                </li>
              )}
              {planHighlights(plan, previous(plan)).map((text) => (
                <li key={text} className="flex gap-2">
                  <CheckIcon className="w-4 h-4 text-primary mt-0.5 flex-none" />
                  {text}
                </li>
              ))}
            </ul>
            <div className="mt-5">{button(plan)}</div>
          </div>
        ))}
      </div>

      {more.map((plan) => (
        <div
          key={plan.id}
          className="rounded-xl border border-dashed px-5 py-4 flex flex-col md:flex-row md:items-center gap-3 md:gap-6"
        >
          <div className="flex-none">
            <div className="text-xs text-muted-foreground">Need more?</div>
            <div className="flex items-center gap-2">
              <h3 className="font-semibold">{plan.title}</h3>
              {badge?.(plan)}
            </div>
            <PlanPrice
              plan={plan}
              interval={interval}
              className="text-xl font-bold"
              suffix="/ month"
            />
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-1.5 flex-1">
            {[
              [shortNumber(plan.feature_usage_credits_per_month), "credits"],
              [shortNumber(plan.feature_max_guilds), "servers"],
              [shortNumber(plan.feature_max_variables), "variables"],
              [
                shortNumber(plan.feature_max_event_listeners),
                "event listeners",
              ],
            ].map(([value, label]) => (
              <div
                key={label}
                className="rounded-lg bg-muted/50 px-1 py-2 text-center"
              >
                <div className="text-lg font-bold">{value}</div>
                <div className="text-xs text-muted-foreground">{label}</div>
              </div>
            ))}
          </div>
          <div className="md:w-40 flex-none">{button(plan)}</div>
        </div>
      ))}
    </div>
  );
}
