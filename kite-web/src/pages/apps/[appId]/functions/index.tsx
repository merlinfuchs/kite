import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import FunctionList from "@/components/app/FunctionList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Functions",
  },
];

export default function AppFunctionsPage() {
  return (
    <AppLayout title="Functions" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Functions</h1>
        <p className="text-muted-foreground text-sm">
          Build blocks you need in several places once, and call them from your
          commands, event listeners and message templates.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/function`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <FunctionList />
    </AppLayout>
  );
}

AppFunctionsPage.getLayout = getAppShellLayout;
