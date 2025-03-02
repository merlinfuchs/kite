import AppLayout from "@/components/app/AppLayout";
import VariableList from "@/components/app/VariableList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
const breadcrumbs = [
  {
    label: "Shared Variables",
  },
];

export default function AppVariablesPage() {
  return (
    <AppLayout title="Shared Variables" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Shared Variables
        </h1>
        <p className="text-muted-foreground text-sm">
          Manage shared variables in your app. Shared variables are key-value
          pairs that can be used to store data across commands, events, and
          more.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/variable`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <VariableList />
    </AppLayout>
  );
}
