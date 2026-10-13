import { App } from "@/lib/types/wire.gen";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { BotIcon } from "lucide-react";
import Link from "next/link";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";

export default function AppListEntry({ app }: { app: App }) {
  return (
    <Card className="flex flex-col sm:flex-row sm:items-start sm:justify-between">
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <BotIcon className="h-5 w-5 flex-none text-muted-foreground" />
          <div>{app.name}</div>
          {!app.enabled && (
            <Badge variant={app.token_invalid ? "destructive" : "secondary"}>
              {app.token_invalid ? "Token stopped working" : "Stopped"}
            </Badge>
          )}
        </CardTitle>
        <CardDescription className="text-sm">{app.description}</CardDescription>
      </CardHeader>
      <div className="flex flex-none gap-2 px-6 pb-6 sm:p-3">
        {app.token_invalid && (
          <AppTokenUpdateDialog app={app}>
            <Button size="sm">Update token</Button>
          </AppTokenUpdateDialog>
        )}
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]",
              query: { appId: app.id },
            }}
          >
            Open app
          </Link>
        </Button>
      </div>
    </Card>
  );
}
