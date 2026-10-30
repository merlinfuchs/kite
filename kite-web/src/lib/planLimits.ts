import { BillingPlan } from "./types/wire.gen";
import { formatInterval, formatNumber } from "./utils";

// Every limit and feature of a plan, in the order the plan details show them.
// A value of true or false is shown as a check or a dash.
export const planLimits: {
  label: string;
  value: (plan: BillingPlan) => string | boolean;
}[] = [
  {
    label: "Credits / month",
    value: (p) => formatNumber(p.feature_usage_credits_per_month),
  },
  { label: "Servers", value: (p) => formatNumber(p.feature_max_guilds) },
  { label: "Commands", value: (p) => formatNumber(p.feature_max_commands) },
  { label: "Variables", value: (p) => formatNumber(p.feature_max_variables) },
  {
    label: "Message templates",
    value: (p) => formatNumber(p.feature_max_messages),
  },
  {
    label: "Event listeners",
    value: (p) => formatNumber(p.feature_max_event_listeners),
  },
  {
    // 0 means no limit, like every other limit.
    label: "Schedules",
    value: (p) =>
      `${p.feature_max_scheduled_event_listeners || "Unlimited"}, ${every(
        p.feature_min_schedule_interval_seconds
      )}`,
  },
  {
    label: "Webhooks",
    value: (p) => `${p.feature_max_webhook_event_listeners || "Unlimited"}`,
  },
  {
    label: "Collaborators",
    value: (p) => formatNumber(p.feature_max_collaborators),
  },
  {
    // 0 means the plan doesn't include the flow AI.
    label: "AI builder prompts / month",
    value: (p) =>
      p.feature_max_ai_prompts_per_month > 0 &&
      formatNumber(p.feature_max_ai_prompts_per_month),
  },
  {
    label: "Rotating & flow-controlled status",
    value: (p) => p.feature_rotating_status,
  },
  {
    label: "Animated bot avatar & banner",
    value: (p) => p.feature_animated_profile,
  },
  {
    label: "Bot avatar, banner & bio per server",
    value: (p) => p.feature_server_profiles,
  },
  { label: "Priority support", value: (p) => p.feature_priority_support },
];

// planHighlights are what a plan card lists below its credits, servers and
// commands: what the plan adds over the one before it, or for the first plan
// what it includes. They're in order of how much they set a plan apart.
const highlights: {
  // Whether a plan lists it, compared to the plan before it if there is one.
  shown: (plan: BillingPlan, previous?: BillingPlan) => boolean;
  text: (plan: BillingPlan) => string;
}[] = [
  {
    shown: more((p) => Number(p.feature_rotating_status)),
    text: () => "Rotating & flow-controlled status",
  },
  {
    shown: more((p) => Number(p.feature_animated_profile)),
    text: () => "Animated bot avatar & banner",
  },
  {
    shown: more((p) => Number(p.feature_server_profiles)),
    text: () => "Bot avatar, banner & bio per server",
  },
  {
    shown: more((p) => Number(p.feature_priority_support)),
    text: () => "Priority support",
  },
  {
    shown: more((p) => p.feature_max_ai_prompts_per_month),
    text: (p) =>
      `${formatNumber(p.feature_max_ai_prompts_per_month)} AI builder prompts`,
  },
  {
    shown: more((p) => p.feature_max_variables + p.feature_max_messages),
    text: (p) =>
      p.feature_max_variables === p.feature_max_messages
        ? `${formatNumber(
            p.feature_max_variables
          )} variables and message templates`
        : `${formatNumber(p.feature_max_variables)} variables, ${formatNumber(
            p.feature_max_messages
          )} message templates`,
  },
  {
    shown: more((p) => p.feature_max_event_listeners),
    text: (p) =>
      `${formatNumber(p.feature_max_event_listeners)} event listeners`,
  },
  {
    // More schedules or schedules that run more often.
    shown: (p, q) =>
      !q ||
      more(schedules)(p, q) ||
      p.feature_min_schedule_interval_seconds <
        q.feature_min_schedule_interval_seconds,
    text: (p) =>
      `${count(schedules(p), "schedules")}, ${every(
        p.feature_min_schedule_interval_seconds
      )}`,
  },
  {
    shown: more(webhooks),
    text: (p) => count(webhooks(p), "webhooks"),
  },
  {
    shown: more((p) => p.feature_max_collaborators),
    text: (p) => `${formatNumber(p.feature_max_collaborators)} collaborators`,
  },
];

export function planHighlights(plan: BillingPlan, previous?: BillingPlan) {
  return highlights
    .filter((h) => h.shown(plan, previous))
    .map((h) => h.text(plan));
}

// more shows a highlight when the plan has more of it than the plan before,
// or for the first plan when it has any.
function more(value: (plan: BillingPlan) => number) {
  return (plan: BillingPlan, previous?: BillingPlan) =>
    previous ? value(plan) > value(previous) : value(plan) > 0;
}

// 0 means no limit for schedules and webhooks.
function schedules(plan: BillingPlan) {
  return plan.feature_max_scheduled_event_listeners || Infinity;
}

function webhooks(plan: BillingPlan) {
  return plan.feature_max_webhook_event_listeners || Infinity;
}

function count(n: number, things: string) {
  return n === Infinity
    ? `Unlimited ${things}`
    : `${formatNumber(n)} ${things}`;
}

// shortNumber writes big limits like 10k and 2.5k, for the plan cards.
export function shortNumber(n: number) {
  if (n >= 1_000_000) return `${n / 1_000_000}M`;
  if (n >= 1_000) return `${n / 1_000}k`;
  return `${n}`;
}

// every reads "every minute" instead of "every 1 minute".
function every(seconds: number) {
  return seconds === 60 ? "every minute" : `every ${formatInterval(seconds)}`;
}
