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
import { formatDate } from "@/lib/utils";

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
              {formatDate(new Date(subscription.created_at))}
              {subscription.ends_at
                ? ` - ${formatDate(new Date(subscription.ends_at))}`
                : subscription.renews_at
                ? ` - renews at ${formatDate(new Date(subscription.renews_at))}`
                : null}
            </CardDescription>
          </div>
          <div>
            <Badge variant={subscription.active ? "default" : "outline"}>
              {subscription.status_formatted}
            </Badge>
          </div>
        </div>
      </CardHeader>
      {subscription.manageable && (
        <CardFooter className="gap-3">
          <Button variant="outline" onClick={() => updatePaymentMethod()}>
            Update Billing
          </Button>
          <Button variant="outline" onClick={() => openCustomerPortal()}>
            Manage
          </Button>
        </CardFooter>
      )}
    </Card>
  );
}
