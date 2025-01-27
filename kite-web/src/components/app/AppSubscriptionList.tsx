import { useAppSubscriptions } from "@/lib/hooks/api";
import AppSubscriptionListEntry from "./AppSubscriptionListEntry";

export default function AppSubscriptionList() {
  const subscriptions = useAppSubscriptions();

  return (
    <div className="flex flex-col gap-4">
      {subscriptions?.map((sub) => (
        <AppSubscriptionListEntry subscription={sub!} />
      ))}
    </div>
  );
}
