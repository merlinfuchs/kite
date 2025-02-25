import { useCallback } from "react";
import {
  useAppSubscriptionManageMutation,
  useCheckoutCreateMutation,
} from "../api/mutations";
import { useAppId } from "./params";
import { toast } from "sonner";

export function useLemonSqueezyCheckout() {
  const appId = useAppId();
  const checkoutMutation = useCheckoutCreateMutation(appId);

  return useCallback(
    (variantId: string) => {
      checkoutMutation.mutate(
        {
          lemonsqueezy_variant_id: variantId,
        },
        {
          onSuccess(res) {
            if (res.success) {
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
