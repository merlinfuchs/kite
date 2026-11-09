import { useApps } from "@/lib/hooks/api";
import Link from "next/link";
import { ComponentProps } from "react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";

// TemplateAppPickerCard lists the apps of the logged in user, each linking to
// the page that adds the template to it. The button of the Discord preview
// links to it.
export default function TemplateAppPickerCard({
  href,
}: {
  href: (appId: string) => ComponentProps<typeof Link>["href"];
}) {
  const apps = useApps();
  const list = apps?.flatMap((a) => (a ? [a] : []));

  return (
    <Card id="use" className="p-5 scroll-mt-20">
      <div className="font-semibold mb-1">Add to your app</div>
      <p className="text-sm text-muted-foreground mb-4">
        {list && !list.length
          ? "Create an app first, then come back to this page."
          : "Choose the app. You can fill in its parameters and choose what to include there."}
      </p>
      <div className="flex flex-col gap-2">
        {list?.map((app) => (
          <Button key={app.id} variant="outline" className="w-full" asChild>
            <Link href={href(app.id)}>{app.name}</Link>
          </Button>
        ))}
        {list && !list.length && (
          <Button className="w-full" asChild>
            <Link href="/apps">Create an app</Link>
          </Button>
        )}
      </div>
    </Card>
  );
}
