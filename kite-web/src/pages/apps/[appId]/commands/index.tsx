import AppLayout from "@/components/app/AppLayout";
import CommandList from "@/components/app/CommandList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Commands",
  },
];

export default function AppCommandsPage() {
  return (
    <AppLayout title="Commands" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Commands</h1>
        <p className="text-muted-foreground text-sm">
          Create custom commands for your app to let users interact with it.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/command`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <CommandList />
    </AppLayout>
  );
}
