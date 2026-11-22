import { App } from "@/lib/types/wire.gen";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";
import BotAvatar from "./BotAvatar";

export default function AppListEntry({ app }: { app: App }) {
  return (
    <Card className="flex flex-col sm:flex-row sm:items-start sm:justify-between">
      <CardHeader className="flex-row items-center gap-4 space-y-0">
        <BotAvatar
          appId={app.id}
          className="size-12 flex-none rounded-xl"
          fallbackClassName="rounded-xl bg-muted text-muted-foreground"
        />
        <div className="grid min-w-0 gap-1.5">
          <CardTitle className="text-base flex flex-wrap items-center gap-2">
            <div>{app.name}</div>
            {app.bot_type === "shared" && (
              <Badge variant="outline" className="whitespace-nowrap">
                Quick setup
              </Badge>
            )}
            {!app.enabled && (
              <Badge
                variant={app.token_invalid ? "destructive" : "secondary"}
                className="whitespace-nowrap"
              >
                {app.token_invalid ? "Token invalid" : "Stopped"}
              </Badge>
            )}
            {app.enabled && app.no_credits && (
              <Badge variant="secondary" className="whitespace-nowrap">
                Out of credits
              </Badge>
            )}
          </CardTitle>
          <CardDescription className="text-sm">
            {app.description}
          </CardDescription>
        </div>
      </CardHeader>
      <div className="flex flex-none gap-2 px-6 pb-6 sm:p-3">
        {app.token_invalid && (
          <AppTokenUpdateDialog app={app} from="list">
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
