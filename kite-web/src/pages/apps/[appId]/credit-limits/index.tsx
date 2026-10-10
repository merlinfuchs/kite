import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import CreditLimitList from "@/components/app/CreditLimitList";
import CreditLimitUsageTable from "@/components/app/CreditLimitUsageTable";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Credit Limits",
  },
];

export default function AppCreditLimitsPage() {
  return (
    <AppLayout title="Credit Limits" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Credit Limits
        </h1>
        <p className="text-muted-foreground text-sm">
          Limit how many credits a single server or user can use per day or
          month, so one busy server or spamming user can&apos;t use up all of
          your app&apos;s credits. Changes take up to a minute to apply.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-limits`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <div className="flex flex-col space-y-8">
        <CreditLimitList />
        <CreditLimitUsageTable />
      </div>
    </AppLayout>
  );
}

AppCreditLimitsPage.getLayout = getAppShellLayout;
