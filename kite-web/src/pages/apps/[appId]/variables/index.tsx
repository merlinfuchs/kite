import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Stored Variables",
  },
];

export default function AppVariablesPage() {
  return (
    <AppLayout title="Stored Variables" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Stored Variables
        </h1>
        <p className="text-muted-foreground text-sm">
          Manage stored variables in your app. Stored variables are key-value
          pairs that can be used to store data across commands, events, and
          more.
        </p>
      </div>
      <Separator className="my-8" />
      <AppEmptyPlaceholder
        title="Under construction"
        description="This feature is not yet available. Please check back later."
      />
      {/* <VariableList /> */}
    </AppLayout>
  );
}
