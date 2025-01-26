import { useCheckoutCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { Button } from "../ui/button";
import { useCallback } from "react";
import { toast } from "sonner";
import { useHookedTheme } from "@/lib/hooks/theme";

export function BillingCheckoutButton() {
  const appId = useAppId();
  const checkoutMutation = useCheckoutCreateMutation(appId);

  const { theme } = useHookedTheme();

  const handleCheckout = useCallback(() => {
    checkoutMutation.mutate(
      {},
      {
        onSuccess(res) {
          if (res.success) {
            (window as any).LemonSqueezy.Url.Open(res.data.url);
          } else {
            toast.error("Failed to create checkout");
          }
        },
      }
    );
  }, [checkoutMutation, theme]);

  return (
    <div>
      <Button onClick={handleCheckout}>Checkout</Button>
    </div>
  );
}
