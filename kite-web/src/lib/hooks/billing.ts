import { useCallback } from "react";
import {
  CheckoutEventNames,
  initializePaddle,
  Paddle,
} from "@paddle/paddle-js";
import {
  useAppSubscriptionManageMutation,
  useAppSubscriptionPlanUpdateMutation,
  useCheckoutCreateMutation,
} from "../api/mutations";
import { useAppId } from "./params";
import { toast } from "sonner";
import { track } from "../analytics";
import {
  BillingCheckoutResponse,
  BillingPlan,
  Subscription,
} from "../types/wire.gen";
import { BillingInterval } from "../billing";

// useCheckout opens the checkout of the plan, of Lemon Squeezy or Paddle,
// whichever the service sells new subscriptions through.
export function useCheckout() {
  const appId = useAppId();
  const checkoutMutation = useCheckoutCreateMutation(appId);

  return useCallback(
    (plan: BillingPlan, interval: BillingInterval) => {
      const properties = { plan: plan.id, interval };

      checkoutMutation.mutate(
        { plan_id: plan.id, interval },
        {
          onSuccess(res) {
            if (!res.success) {
              toast.error(
                `Failed to create checkout: ${res.error.message} ${res.error.code}`
              );
              return;
            }

            track("checkout_started", properties);
            if (res.data.provider === "paddle") {
              openPaddleCheckout(res.data, properties);
            } else {
              openLemonSqueezyCheckout(res.data.url!, properties);
            }
          },
        }
      );
    },
    [checkoutMutation]
  );
}

function openLemonSqueezyCheckout(
  url: string,
  properties: Record<string, string>
) {
  const lemonSqueezy = (window as any).LemonSqueezy;
  // Lemon.js has no event for closing the checkout, so abandoned checkouts
  // are the started ones that never complete.
  lemonSqueezy.Setup({
    eventHandler: (event: { event: string }) => {
      if (event.event === "Checkout.Success") {
        track("checkout_completed", properties);
      }
    },
  });
  lemonSqueezy.Url.Open(url);
}

let paddle: Promise<Paddle | undefined> | undefined;
// The properties of the open checkout, for the events of Paddle.js, which
// are set up once for all checkouts.
let paddleCheckoutProperties: Record<string, string> | undefined;

// loadPaddle loads Paddle.js the first time a checkout opens, so pages without
// one don't load it.
function loadPaddle(token: string, sandbox: boolean) {
  paddle ??= initializePaddle({
    token,
    environment: sandbox ? "sandbox" : "production",
    eventCallback(event) {
      if (
        event.name === CheckoutEventNames.CHECKOUT_COMPLETED &&
        paddleCheckoutProperties
      ) {
        track("checkout_completed", paddleCheckoutProperties);
      }
    },
  }).catch((err) => {
    // Retried with the next checkout, like after an ad blocker is turned off.
    paddle = undefined;
    throw err;
  });
  return paddle;
}

async function openPaddleCheckout(
  checkout: BillingCheckoutResponse,
  properties: Record<string, string>
) {
  try {
    const instance = await loadPaddle(
      checkout.paddle_client_token!,
      !!checkout.paddle_sandbox
    );
    if (!instance) throw new Error("Paddle.js isn't available");

    paddleCheckoutProperties = properties;
    instance.Checkout.open({
      transactionId: checkout.paddle_transaction_id!,
      customer: checkout.customer_email
        ? { email: checkout.customer_email }
        : undefined,
      settings: { successUrl: checkout.success_url },
    });
  } catch {
    toast.error(
      "The checkout couldn't open. Check that nothing blocks paddle.com and try again."
    );
  }
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
        { plan_id: plan.id, interval },
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

export function useSubscriptionCustomerPortal(subscription: Subscription) {
  const manageMutation = useAppSubscriptionManageMutation(subscription.id);

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

export function useSubscriptionUpdatePaymentMethod(subscription: Subscription) {
  const manageMutation = useAppSubscriptionManageMutation(subscription.id);
  const source = subscription.source;

  return useCallback(() => {
    manageMutation.mutate(undefined, {
      onSuccess(res) {
        if (!res.success) {
          toast.error(
            `Failed to manage subscription: ${res.error.message} ${res.error.code}`
          );
        } else if (source === "paddle") {
          // A page of Paddle's customer portal, not an overlay.
          window.location.href = res.data.update_payment_method_url;
        } else {
          (window as any).LemonSqueezy.Url.Open(
            res.data.update_payment_method_url
          );
        }
      },
    });
  }, [manageMutation, source]);
}
