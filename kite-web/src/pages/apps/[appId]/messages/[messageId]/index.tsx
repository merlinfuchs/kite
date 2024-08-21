import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Message Templates",
    href: "/apps/[appId]/messages",
  },
  {
    label: "Some Message",
  },
];

export default function AppMessagePage() {
  return (
    <AppLayout title="Some Message" breadcrumbs={breadcrumbs}>
      <Separator className="mb-4 -mt-4" />
    </AppLayout>
  );
}
