import { useCallback } from "react";
import {
  useAppSubscriptionManageMutation,
  useAppSubscriptionPlanUpdateMutation,
  useCheckoutCreateMutation,
} from "../api/mutations";
import { useAppId } from "./params";
import { toast } from "sonner";
import { track } from "../analytics";
import { BillingPlan } from "../types/wire.gen";
import { BillingInterval, planVariantId } from "../billing";

export function useLemonSqueezyCheckout() {
  const appId = useAppId();
  const checkoutMutation = useCheckoutCreateMutation(appId);

  return useCallback(
    (plan: BillingPlan, interval: BillingInterval) => {
      checkoutMutation.mutate(
        {
          lemonsqueezy_variant_id: planVariantId(plan, interval),
        },
        {
          onSuccess(res) {
            if (res.success) {
              track("checkout_started", { plan: plan.id, interval });
              (window as any).LemonSqueezy.Url.Open(res.data.url);
            } else {
              toast.error(
                `Failed to create checkout: ${res.error.message} ${res.error.code}`
              );
            }
          },
        }
      );
    },
    [checkoutMutation]
  );
}

// useSubscriptionPlanSwitch moves an existing subscription onto another plan.
// The returned callback is a no-op without a subscription to switch, so callers
// can hold the hook unconditionally.
export function useSubscriptionPlanSwitch(subscriptionId: string | undefined) {
  const appId = useAppId();
  const planUpdateMutation = useAppSubscriptionPlanUpdateMutation(
    appId,
    subscriptionId ?? ""
  );

  return useCallback(
    (plan: BillingPlan, interval: BillingInterval) => {
      if (!subscriptionId) return;

      planUpdateMutation.mutate(
        {
          lemonsqueezy_variant_id: planVariantId(plan, interval),
        },
        {
          onSuccess(res) {
            if (res.success) {
              track("plan_switched", { plan: plan.id, interval });
              toast.success("Your plan has been updated.");
            } else {
              toast.error(
                `Failed to switch plan: ${res.error.message} ${res.error.code}`
              );
            }
          },
        }
      );
    },
    [planUpdateMutation, subscriptionId]
  );
}

export function useLemonSqueezyCustomerPortal(subscriptionId: string) {
  const manageMutation = useAppSubscriptionManageMutation(subscriptionId);

  return useCallback(() => {
    manageMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          window.location.href = res.data.customer_portal_url;
        } else {
          toast.error(
            `Failed to manage subscription: ${res.error.message} ${res.error.code}`
          );
        }
      },
    });
  }, [manageMutation]);
}

export function useLemonSqueezyUpdatePaymentMethod(subscriptionId: string) {
  const manageMutation = useAppSubscriptionManageMutation(subscriptionId);

  return useCallback(() => {
    manageMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          (window as any).LemonSqueezy.Url.Open(
            res.data.update_payment_method_url
          );
        } else {
          toast.error(
            `Failed to manage subscription: ${res.error.message} ${res.error.code}`
          );
        }
      },
    });
  }, [manageMutation]);
}
