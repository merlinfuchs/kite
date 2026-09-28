import AppLayout from "@/components/app/AppLayout";
import CommandList from "@/components/app/CommandList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Context Menus",
  },
];

export default function AppContextMenusPage() {
  return (
    <AppLayout title="Context Menus" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Context Menus</h1>
        <p className="text-muted-foreground text-sm">
          Create user and message context menu commands that appear when
          right-clicking a user or message.{" "}
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
      <CommandList kind="context_menu" />
    </AppLayout>
  );
}
