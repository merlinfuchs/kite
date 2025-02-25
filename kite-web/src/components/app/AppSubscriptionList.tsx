import { useAppSubscriptions } from "@/lib/hooks/api";
import AppSubscriptionListEntry from "./AppSubscriptionListEntry";

export default function AppSubscriptionList() {
  const subscriptions = useAppSubscriptions();

  if (subscriptions?.length === 0) {
    return null;
  }

  return (
    <div>
      <h2 className="text-lg font-semibold md:text-2xl mt-32 mb-6">
        Subscription History
      </h2>
      <div className="flex flex-col gap-4">
        {subscriptions?.map((sub) => (
          <AppSubscriptionListEntry key={sub!.id} subscription={sub!} />
        ))}
      </div>
    </div>
  );
}
