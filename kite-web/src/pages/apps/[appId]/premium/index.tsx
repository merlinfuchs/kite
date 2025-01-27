import AppLayout from "@/components/app/AppLayout";
import { breadcrumbs } from "../commands";
import { Separator } from "@/components/ui/separator";
import { BillingCheckoutButton } from "@/components/app/BillingCheckoutButton";

export default function AppPremiumPage() {
  return (
    <AppLayout title="Premium" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Premium</h1>
        <p className="text-muted-foreground text-sm">
          Manage your app&apos;s access to premium features and subscriptions.
        </p>
      </div>
      <Separator className="my-8" />

      <BillingCheckoutButton />
    </AppLayout>
  );
}
