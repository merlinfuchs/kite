import { Subscription } from "@/lib/types/wire.gen";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  useSubscriptionCustomerPortal,
  useSubscriptionMove,
  useSubscriptionUpdatePaymentMethod,
} from "@/lib/hooks/billing";
import { formatDate } from "@/lib/utils";

export default function AppSubscriptionListEntry({
  subscription,
}: {
  subscription: Subscription;
}) {
  const updatePaymentMethod = useSubscriptionUpdatePaymentMethod(subscription);
  const openCustomerPortal = useSubscriptionCustomerPortal(subscription.id);
  const move = useSubscriptionMove(subscription);

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
      {subscription.movable && subscription.renews_at && (
        <CardContent className="text-sm text-muted-foreground">
          Lemon Squeezy, which bills this subscription, shuts down on March 31,
          2027. Move it to Paddle to keep your plan after that. You pay nothing
          today, and your next payment stays on{" "}
          {formatDate(new Date(subscription.renews_at))}. We cancel the Lemon
          Squeezy subscription for you.
        </CardContent>
      )}
      {subscription.manageable && (
        <CardFooter className="gap-3 flex-wrap">
          {subscription.movable && (
            <Button onClick={() => move()}>Move to Paddle</Button>
          )}
          <Button variant="outline" onClick={() => updatePaymentMethod()}>
            Update billing
          </Button>
          <Button variant="outline" onClick={() => openCustomerPortal()}>
            Manage
          </Button>
        </CardFooter>
      )}
    </Card>
  );
}
