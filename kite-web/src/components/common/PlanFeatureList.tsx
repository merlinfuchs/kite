import { BillingPlan } from "@/lib/types/wire.gen";
import { formatInterval, formatNumber } from "@/lib/utils";
import { CheckIcon } from "lucide-react";

// The limits and features of a plan, on the premium page and the homepage.
export default function PlanFeatureList({ plan }: { plan: BillingPlan }) {
  return (
    <div className="space-y-4">
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {formatNumber(plan.feature_max_collaborators)} Collaborator
          {plan.feature_max_collaborators === 1 ? "" : "s"}
        </h3>
      </span>
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {formatNumber(plan.feature_usage_credits_per_month)} Credits / month
        </h3>
      </span>
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {formatNumber(plan.feature_max_guilds)} Servers
        </h3>
      </span>
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {formatNumber(plan.feature_max_commands)} Commands & Variables
        </h3>
      </span>
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {formatNumber(plan.feature_max_event_listeners)} Event Listeners
        </h3>
      </span>
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {/* 0 means no limit, like every other limit. */}
          {plan.feature_max_scheduled_event_listeners || "Unlimited"} Schedules,
          every {formatInterval(plan.feature_min_schedule_interval_seconds)}
        </h3>
      </span>
      {/* 0 means the plan doesn't include the flow AI. */}
      {plan.feature_max_ai_prompts_per_month > 0 && (
        <span className="flex">
          <CheckIcon className="flex-none text-primary" />{" "}
          <h3 className="ml-2">
            {formatNumber(plan.feature_max_ai_prompts_per_month)} AI builder
            prompts / month
          </h3>
        </span>
      )}
      {plan.feature_rotating_status && (
        <span className="flex">
          <CheckIcon className="flex-none text-primary" />{" "}
          <h3 className="ml-2">Rotating & flow-controlled status</h3>
        </span>
      )}
      <span className="flex">
        <CheckIcon className="flex-none text-primary" />{" "}
        <h3 className="ml-2">
          {plan.feature_priority_support
            ? "Priority support"
            : "Community support"}
        </h3>
      </span>
    </div>
  );
}
