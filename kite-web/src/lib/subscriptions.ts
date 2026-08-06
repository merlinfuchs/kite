import { Subscription } from "./types/wire.gen";

// LemonSqueezy subscription statuses that still grant (or will grant) access.
// Anything else - "cancelled", "expired", "unpaid", "paused" - leaves the app
// without an entitlement, so the plan must be purchasable again.
const ACTIVE_STATUSES = ["on_trial", "active", "past_due"];

export function isSubscriptionActive(subscription: Subscription): boolean {
  return ACTIVE_STATUSES.includes(subscription.status);
}
