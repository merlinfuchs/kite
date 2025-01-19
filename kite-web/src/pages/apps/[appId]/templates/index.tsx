import AppLayout from "@/components/app/AppLayout";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Templates",
  },
];

export default function AppTemplatesPage() {
  return (
    <AppLayout title="App Templates" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Templates</h1>
        <p className="text-muted-foreground text-sm">
          Select any of the templates below to get started. Templates help you
          build your app faster and can contain commands, event listeners,
          message templates, and more.
        </p>
      </div>
      <Separator className="my-8" />
      <TemplateList />
    </AppLayout>
  );
}
