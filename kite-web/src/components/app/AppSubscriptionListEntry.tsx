import { Subscription } from "@/lib/types/wire.gen";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  useLemonSqueezyCustomerPortal,
  useLemonSqueezyUpdatePaymentMethod,
} from "@/lib/hooks/lemonsqueezy";

export default function AppSubscriptionListEntry({
  subscription,
}: {
  subscription: Subscription;
}) {
  const updatePaymentMethod = useLemonSqueezyUpdatePaymentMethod(
    subscription.id
  );
  const openCustomerPortal = useLemonSqueezyCustomerPortal(subscription.id);

  return (
    <Card>
      <CardHeader>
        <div className="flex justify-between">
          <div>
            <CardTitle className="text-xl mb-1">
              {subscription.display_name}
            </CardTitle>
            <CardDescription>
              {new Date(subscription.created_at).toLocaleDateString()}
              {subscription.ends_at
                ? ` - ${new Date(subscription.ends_at).toLocaleDateString()}`
                : " - renews at " +
                  new Date(subscription.renews_at).toLocaleDateString()}
            </CardDescription>
          </div>
          <div>
            <Badge
              variant={subscription.status !== "ended" ? "default" : "outline"}
            >
              {subscription.status_formatted}
            </Badge>
          </div>
        </div>
      </CardHeader>
      <CardFooter className="gap-3">
        {subscription.manageable && (
          <>
            <Button variant="outline" onClick={() => updatePaymentMethod()}>
              Update Billing
            </Button>
            <Button variant="outline" onClick={() => openCustomerPortal()}>
              Manage
            </Button>
          </>
        )}
      </CardFooter>
    </Card>
  );
}
