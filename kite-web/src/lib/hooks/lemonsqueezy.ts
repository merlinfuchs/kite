import { useCallback } from "react";
import { useCheckoutCreateMutation } from "../api/mutations";
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
