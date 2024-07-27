import { App } from "@/lib/types/wire.gen";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { PackageIcon } from "lucide-react";
import Link from "next/link";

export default function AppListEntry({ app }: { app: App }) {
  return (
    <Card>
      <Button size="sm" variant="outline" className="float-right m-3" asChild>
        <Link
          href={{
            pathname: "/apps/[appId]",
            query: { appId: app.id },
          }}
        >
          Open app
        </Link>
      </Button>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <PackageIcon className="h-5 w-5 text-muted-foreground" />
          <div>{app.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">{app.description}</CardDescription>
      </CardHeader>
    </Card>
  );
}
