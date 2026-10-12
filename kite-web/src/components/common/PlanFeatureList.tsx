import { BillingPlan } from "@/lib/types/wire.gen";
import { formatInterval, formatNumber } from "@/lib/utils";
import { CheckIcon } from "lucide-react";
import { ReactNode } from "react";

// The limits and features of a plan, on the premium page and the homepage.
export default function PlanFeatureList({ plan }: { plan: BillingPlan }) {
  return (
    <ul className="space-y-4">
      <Feature>
        {formatNumber(plan.feature_max_collaborators)} Collaborator
        {plan.feature_max_collaborators === 1 ? "" : "s"}
      </Feature>
      <Feature>
        {formatNumber(plan.feature_usage_credits_per_month)} Credits / month
      </Feature>
      <Feature>{formatNumber(plan.feature_max_guilds)} Servers</Feature>
      <Feature>
        {formatNumber(plan.feature_max_commands)} Commands & Variables
      </Feature>
      <Feature>
        {formatNumber(plan.feature_max_event_listeners)} Event Listeners
      </Feature>
      <Feature>
        {/* 0 means no limit, like every other limit. */}
        {plan.feature_max_scheduled_event_listeners || "Unlimited"} Schedules,
        every {formatInterval(plan.feature_min_schedule_interval_seconds)}
      </Feature>
      {/* 0 means the plan doesn't include the flow AI. */}
      {plan.feature_max_ai_prompts_per_month > 0 && (
        <Feature>
          {formatNumber(plan.feature_max_ai_prompts_per_month)} AI builder
          prompts / month
        </Feature>
      )}
      {plan.feature_rotating_status && (
        <Feature>Rotating & flow-controlled status</Feature>
      )}
      <Feature>
        {plan.feature_priority_support
          ? "Priority support"
          : "Community support"}
      </Feature>
    </ul>
  );
}

function Feature({ children }: { children: ReactNode }) {
  return (
    <li className="flex gap-2">
      <CheckIcon className="flex-none text-primary" />
      <span>{children}</span>
    </li>
  );
}
