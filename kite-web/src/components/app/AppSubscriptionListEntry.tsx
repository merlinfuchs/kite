import { Subscription } from "@/lib/types/wire.gen";
import { Card, CardHeader, CardTitle } from "../ui/card";

export default function AppSubscriptionListEntry({
  subscription,
}: {
  subscription: Subscription;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{subscription.display_name}</CardTitle>
      </CardHeader>
    </Card>
  );
}
